package store

import (
	"database/sql"
	"time"

	"github.com/caleb/sports-catalog/models"
)

type OrderStore struct {
	db *sql.DB
}

func NewOrderStore(db *sql.DB) *OrderStore {
	return &OrderStore{db: db}
}

func (s *OrderStore) Create(o *models.Order) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	res, err := tx.Exec(`
		INSERT INTO orders (stripe_session_id, status, total_cents, currency, customer_email, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		o.StripeSessionID, o.Status, o.TotalCents, o.Currency, o.CustomerEmail, now)
	if err != nil {
		return 0, err
	}
	orderID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, item := range o.Items {
		if _, err := tx.Exec(`
			INSERT INTO order_items (order_id, product_id, sku, name, quantity, unit_price_cents)
			VALUES (?, ?, ?, ?, ?, ?)`,
			orderID, item.ProductID, item.SKU, item.Name, item.Quantity, item.UnitPriceCents); err != nil {
			return 0, err
		}
	}

	return orderID, tx.Commit()
}

// UpdateStatusBySessionID is called from the Stripe webhook handler once
// checkout.session.completed (or a failure event) comes in.
func (s *OrderStore) UpdateStatusBySessionID(sessionID string, status models.OrderStatus) error {
	_, err := s.db.Exec(`UPDATE orders SET status = ? WHERE stripe_session_id = ?`, status, sessionID)
	return err
}

func (s *OrderStore) GetBySessionID(sessionID string) (*models.Order, error) {
	var o models.Order
	err := s.db.QueryRow(`
		SELECT id, stripe_session_id, status, total_cents, currency, customer_email, created_at
		FROM orders WHERE stripe_session_id = ?`, sessionID,
	).Scan(&o.ID, &o.StripeSessionID, &o.Status, &o.TotalCents, &o.Currency, &o.CustomerEmail, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &o, nil
}
