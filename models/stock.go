package models

import "time"

// StockItem tracks physical inventory for a product at a given warehouse/location.
// A product can have multiple stock rows (one per warehouse); the catalogue
// itself never stores quantity. This mirrors keeping "what it is" separate
// from "how many/where", which scales better once you add multiple locations
// or event-specific stock pools (e.g. a pop-up stall vs main warehouse).
type StockItem struct {
	ID        int64     `json:"id" db:"id"`
	ProductID int64     `json:"product_id" db:"product_id"`
	Warehouse string    `json:"warehouse" db:"warehouse"`
	Quantity  int       `json:"quantity" db:"quantity"`
	ReorderAt int       `json:"reorder_at" db:"reorder_at"` // trigger a low-stock flag below this
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// StockAdjustment is the input shape for increment/decrement operations,
// e.g. receiving a shipment or fulfilling an order.
type StockAdjustment struct {
	ProductID int64  `json:"product_id"`
	Warehouse string `json:"warehouse"`
	Delta     int    `json:"delta"` // positive = restock, negative = deduct
	Reason    string `json:"reason"`
}
