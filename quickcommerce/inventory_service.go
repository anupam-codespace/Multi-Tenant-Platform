package quickcommerce

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrInsufficientStock   = errors.New("insufficient inventory available in dark store")
	ErrReservationNotFound = errors.New("reservation not found or already expired")
	ErrDuplicateRequest    = errors.New("duplicate request detected with active idempotency key")
)

// StockReservation represents an atomic hold on dark store inventory during checkout.
// Prevents flash-sale race conditions where stock can oversell.
type StockReservation struct {
	ReservationID string    `json:"reservation_id"`
	StoreID       int64     `json:"store_id"`
	ItemID        int64     `json:"item_id"`
	Quantity      int       `json:"quantity"`
	UserID        int64     `json:"user_id"`
	ExpiresAt     time.Time `json:"expires_at"`
	IsCommitted   bool      `json:"is_committed"`
}

// InventoryManager coordinates atomic stock decrements, locks, and TTL-based rollbacks.
type InventoryManager struct {
	mu           sync.RWMutex
	storeStock   map[string]int // key: "storeID:itemID" -> available stock
	reservations map[string]*StockReservation
	ttl          time.Duration
}

// NewInventoryManager initializes an in-memory high-throughput dark store inventory manager.
func NewInventoryManager(ttl time.Duration) *InventoryManager {
	im := &InventoryManager{
		storeStock:   make(map[string]int),
		reservations: make(map[string]*StockReservation),
		ttl:          ttl,
	}

	// Background worker: sweeps expired reservations and restores stock
	go im.startTTLExpirationWorker(context.Background(), 10*time.Second)

	return im
}

// SeedStock sets the initial stock quantity for a store's item.
func (im *InventoryManager) SeedStock(storeID, itemID int64, quantity int) {
	im.mu.Lock()
	defer im.mu.Unlock()
	key := fmt.Sprintf("%d:%d", storeID, itemID)
	im.storeStock[key] = quantity
}

// GetAvailableStock retrieves real-time available stock excluding active reservations.
func (im *InventoryManager) GetAvailableStock(storeID, itemID int64) int {
	im.mu.RLock()
	defer im.mu.RUnlock()
	key := fmt.Sprintf("%d:%d", storeID, itemID)
	return im.storeStock[key]
}

// ReserveStock atomically holds inventory for a checkout session with a TTL.
func (im *InventoryManager) ReserveStock(ctx context.Context, reservationID string, storeID, itemID, userID int64, quantity int) (*StockReservation, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	key := fmt.Sprintf("%d:%d", storeID, itemID)
	currentStock := im.storeStock[key]

	if currentStock < quantity {
		return nil, ErrInsufficientStock
	}

	// Atomically decrement available stock
	im.storeStock[key] = currentStock - quantity

	res := &StockReservation{
		ReservationID: reservationID,
		StoreID:       storeID,
		ItemID:        itemID,
		Quantity:      quantity,
		UserID:        userID,
		ExpiresAt:     time.Now().Add(im.ttl),
		IsCommitted:   false,
	}

	im.reservations[reservationID] = res
	return res, nil
}

// CommitReservation permanently consumes reserved stock when payment succeeds.
func (im *InventoryManager) CommitReservation(reservationID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	res, exists := im.reservations[reservationID]
	if !exists || time.Now().After(res.ExpiresAt) {
		return ErrReservationNotFound
	}

	res.IsCommitted = true
	delete(im.reservations, reservationID)
	return nil
}

// ReleaseReservation cancels an active hold (e.g., payment cancelled or user discarded cart).
func (im *InventoryManager) ReleaseReservation(reservationID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	res, exists := im.reservations[reservationID]
	if !exists {
		return ErrReservationNotFound
	}

	if !res.IsCommitted {
		key := fmt.Sprintf("%d:%d", res.StoreID, res.ItemID)
		im.storeStock[key] += res.Quantity
	}

	delete(im.reservations, reservationID)
	return nil
}

// startTTLExpirationWorker automatically frees held stock if checkout is abandoned.
func (im *InventoryManager) startTTLExpirationWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			im.mu.Lock()
			now := time.Now()
			for id, res := range im.reservations {
				if !res.IsCommitted && now.After(res.ExpiresAt) {
					key := fmt.Sprintf("%d:%d", res.StoreID, res.ItemID)
					im.storeStock[key] += res.Quantity
					delete(im.reservations, id)
				}
			}
			im.mu.Unlock()
		}
	}
}
