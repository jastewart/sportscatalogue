package models

import "time"

type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderPaid      OrderStatus = "paid"
	OrderFailed    OrderStatus = "failed"
	OrderCancelled OrderStatus = "cancelled"
)

type OrderItem struct {
	ProductID      int64  `json:"product_id"`
	SKU            string `json:"sku"`
	Name           string `json:"name"`
	Quantity       int    `json:"quantity"`
	UnitPriceCents int64  `json:"unit_price_cents"`
}

type Order struct {
	ID              int64       `json:"id" db:"id"`
	StripeSessionID string      `json:"stripe_session_id" db:"stripe_session_id"`
	Status          OrderStatus `json:"status" db:"status"`
	TotalCents      int64       `json:"total_cents" db:"total_cents"`
	Currency        string      `json:"currency" db:"currency"`
	CustomerEmail   string      `json:"customer_email" db:"customer_email"`
	Items           []OrderItem `json:"items" db:"-"`
	CreatedAt       time.Time   `json:"created_at" db:"created_at"`
}

// CheckoutRequest is what the client posts to start a mock Stripe checkout.
type CheckoutRequest struct {
	CustomerEmail string `json:"customer_email"`
	Items         []struct {
		ProductID int64 `json:"product_id"`
		Quantity  int   `json:"quantity"`
	} `json:"items"`
}
