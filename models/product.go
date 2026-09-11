package models

import "time"

// CategoryType is the top-level split between sports equipment and everything else
// in the catalogue. This is the field that lets you query/report on the two groups
// separately even though they live in the same products table.
type CategoryType string

const (
	CategorySportsEquipment CategoryType = "sports_equipment"
	CategoryApparel         CategoryType = "apparel"
	CategoryAccessory       CategoryType = "accessory"
	CategoryGeneral         CategoryType = "general"
)

func (c CategoryType) Valid() bool {
	switch c {
	case CategorySportsEquipment, CategoryApparel, CategoryAccessory, CategoryGeneral:
		return true
	}
	return false
}

// Product is a catalogue entry. It intentionally has NO quantity field —
// "how many do we have" lives in StockItem (models/stock.go). This keeps
// catalogue metadata (name, price, description) separate from inventory
// state, which changes far more often and can live per-warehouse.
type Product struct {
	ID           int64        `json:"id" db:"id"`
	SKU          string       `json:"sku" db:"sku"`
	Name         string       `json:"name" db:"name"`
	Description  string       `json:"description" db:"description"`
	CategoryType CategoryType `json:"category_type" db:"category_type"`
	Subcategory  string       `json:"subcategory" db:"subcategory"` // e.g. "Basketball", "Running Shoes"
	PriceCents   int64        `json:"price_cents" db:"price_cents"`
	Currency     string       `json:"currency" db:"currency"`
	ImageURL     string       `json:"image_url" db:"image_url"`
	Active       bool         `json:"active" db:"active"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" db:"updated_at"`
}

// ProductWithStock is a convenience read-model joining catalogue + total stock,
// used by list endpoints so the client doesn't need two round trips.
type ProductWithStock struct {
	Product
	TotalQuantity int `json:"total_quantity" db:"total_quantity"`
}
