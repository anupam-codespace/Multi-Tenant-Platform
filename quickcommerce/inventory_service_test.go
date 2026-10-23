package quickcommerce

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestInventoryManager_ConcurrentReservations(t *testing.T) {
	im := NewInventoryManager(2 * time.Minute)
	storeID := int64(101)
	itemID := int64(505)
	initialStock := 50

	im.SeedStock(storeID, itemID, initialStock)

	// Simulate 100 concurrent checkout attempts trying to reserve 1 item each
	// Exactly 50 should succeed, and exactly 50 should fail with ErrInsufficientStock
	concurrency := 100
	var wg sync.WaitGroup
	wg.Add(concurrency)

	var successCount int64
	var mu sync.Mutex

	for i := 0; i < concurrency; i++ {
		go func(idx int) {
			defer wg.Done()
			resID := fmt.Sprintf("res-%d", idx)
			_, err := im.ReserveStock(context.Background(), resID, storeID, itemID, int64(idx), 1)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	if successCount != 50 {
		t.Fatalf("expected exactly 50 successful reservations, got %d", successCount)
	}

	remaining := im.GetAvailableStock(storeID, itemID)
	if remaining != 0 {
		t.Fatalf("expected remaining stock to be 0, got %d", remaining)
	}
}

func TestInventoryManager_ReleaseReservation(t *testing.T) {
	im := NewInventoryManager(1 * time.Minute)
	storeID := int64(1)
	itemID := int64(2)
	im.SeedStock(storeID, itemID, 10)

	resID := "res-release-test"
	_, err := im.ReserveStock(context.Background(), resID, storeID, itemID, 42, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if im.GetAvailableStock(storeID, itemID) != 6 {
		t.Fatalf("expected 6 stock after reservation, got %d", im.GetAvailableStock(storeID, itemID))
	}

	// Release reservation (cart discarded)
	err = im.ReleaseReservation(resID)
	if err != nil {
		t.Fatalf("failed to release reservation: %v", err)
	}

	if im.GetAvailableStock(storeID, itemID) != 10 {
		t.Fatalf("expected stock restored to 10, got %d", im.GetAvailableStock(storeID, itemID))
	}
}
