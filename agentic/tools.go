package agentic

import (
	"context"
	"fmt"
	"strings"
)

// CatalogItem represents an item available in an Instamart dark store
type CatalogItem struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Brand       string  `json:"brand"`
	Price       float64 `json:"price"`
	Unit        string  `json:"unit"`
	InStock     bool    `json:"in_stock"`
	StockCount  int     `json:"stock_count"`
	StoreID     int64   `json:"store_id"`
}

// DarkStore represents an Instamart micro-fulfillment warehouse
type DarkStore struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Location  string `json:"location"`
	Pincode   string `json:"pincode"`
	IsActive  bool   `json:"is_active"`
	AvgETA    string `json:"avg_eta"`
}

// AgentTools provides executable tools for the Agentic AI loop
type AgentTools interface {
	FindNearestDarkStore(ctx context.Context, pincode string) (*DarkStore, error)
	SearchCatalog(ctx context.Context, storeID int64, query string) ([]CatalogItem, error)
	GetSubstitutions(ctx context.Context, outOfStockItem CatalogItem) ([]CatalogItem, error)
}

// InMemoryStoreCatalog implements mockable and fast catalog tools for the agent
type InMemoryStoreCatalog struct {
	stores   []DarkStore
	products []CatalogItem
}

func NewInMemoryStoreCatalog() *InMemoryStoreCatalog {
	return &InMemoryStoreCatalog{
		stores: []DarkStore{
			{ID: 101, Name: "Instamart Pod - Whitefield", Location: "Whitefield, Bangalore", Pincode: "560066", IsActive: true, AvgETA: "9 mins"},
			{ID: 102, Name: "Instamart Pod - Indiranagar", Location: "Indiranagar, Bangalore", Pincode: "560038", IsActive: true, AvgETA: "11 mins"},
			{ID: 103, Name: "Instamart Pod - Koramangala", Location: "Koramangala, Bangalore", Pincode: "560034", IsActive: true, AvgETA: "8 mins"},
		},
		products: []CatalogItem{
			{ID: 1, Name: "Fresh Paneer", Category: "Dairy", Brand: "Nandini", Price: 110.0, Unit: "200g", InStock: true, StockCount: 25, StoreID: 101},
			{ID: 2, Name: "Organic Paneer", Category: "Dairy", Brand: "Milky Mist", Price: 135.0, Unit: "200g", InStock: true, StockCount: 15, StoreID: 101},
			{ID: 3, Name: "Basmati Rice Supreme", Category: "Grains", Brand: "India Gate", Price: 165.0, Unit: "1kg", InStock: true, StockCount: 40, StoreID: 101},
			{ID: 4, Name: "Basmati Rice Royal", Category: "Grains", Brand: "Daawat", Price: 175.0, Unit: "1kg", InStock: false, StockCount: 0, StoreID: 101},
			{ID: 5, Name: "Butter", Category: "Dairy", Brand: "Amul", Price: 58.0, Unit: "100g", InStock: true, StockCount: 60, StoreID: 101},
			{ID: 6, Name: "Tomato Puree", Category: "Pantry", Brand: "Dabur Hommade", Price: 35.0, Unit: "200g", InStock: true, StockCount: 30, StoreID: 101},
			{ID: 7, Name: "Garam Masala", Category: "Spices", Brand: "Everest", Price: 42.0, Unit: "50g", InStock: true, StockCount: 50, StoreID: 101},
			{ID: 8, Name: "Fresh Coriander", Category: "Vegetables", Brand: "Instamart Fresh", Price: 15.0, Unit: "1 bunch", InStock: true, StockCount: 100, StoreID: 101},
		},
	}
}

func (c *InMemoryStoreCatalog) FindNearestDarkStore(ctx context.Context, pincode string) (*DarkStore, error) {
	for _, store := range c.stores {
		if store.Pincode == pincode {
			return &store, nil
		}
	}
	// Default to Whitefield primary pod
	return &c.stores[0], nil
}

func (c *InMemoryStoreCatalog) SearchCatalog(ctx context.Context, storeID int64, query string) ([]CatalogItem, error) {
	lowerQ := strings.ToLower(query)
	var matches []CatalogItem

	for _, item := range c.products {
		if item.StoreID == storeID || storeID == 0 {
			if strings.Contains(strings.ToLower(item.Name), lowerQ) ||
				strings.Contains(strings.ToLower(item.Category), lowerQ) ||
				strings.Contains(strings.ToLower(item.Brand), lowerQ) {
				matches = append(matches, item)
			}
		}
	}
	return matches, nil
}

func (c *InMemoryStoreCatalog) GetSubstitutions(ctx context.Context, outOfStockItem CatalogItem) ([]CatalogItem, error) {
	var subs []CatalogItem
	for _, item := range c.products {
		if item.ID != outOfStockItem.ID && item.Category == outOfStockItem.Category && item.InStock {
			subs = append(subs, item)
		}
	}
	if len(subs) == 0 {
		return nil, fmt.Errorf("no direct substitution found for category %s", outOfStockItem.Category)
	}
	return subs, nil
}
