package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/caleb/sports-catalog/models"
)

var ErrInsufficientStock = errors.New("insufficient stock")

type StockStore struct {
	db *sql.DB
}

func NewStockStore(db *sql.DB) *StockStore {
	return &StockStore{db: db}
}

// Adjust applies a delta (positive to restock, negative to deduct) inside a
// transaction, guarding against a deduction driving quantity below zero.
// This is the function order fulfillment and manual restocks both call.
func (s *StockStore) Adjust(adj models.StockAdjustment) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	warehouse := adj.Warehouse
	if warehouse == "" {
		warehouse = "main"
	}

	// Make sure a row exists for this product/warehouse pair.
	if _, err := tx.Exec(`
		INSERT INTO stock_items (product_id, warehouse, quantity, updated_at)
		VALUES (?, ?, 0, ?)
		ON CONFLICT(product_id, warehouse) DO NOTHING`,
		adj.ProductID, warehouse, time.Now().UTC()); err != nil {
		return err
	}

	if adj.Delta < 0 {
		var current int
		if err := tx.QueryRow(`SELECT quantity FROM stock_items WHERE product_id = ? AND warehouse = ?`,
			adj.ProductID, warehouse).Scan(&current); err != nil {
			return err
		}
		if current+adj.Delta < 0 {
			return fmt.Errorf("%w: have %d, need %d", ErrInsufficientStock, current, -adj.Delta)
		}
	}

	if _, err := tx.Exec(`
		UPDATE stock_items SET quantity = quantity + ?, updated_at = ?
		WHERE product_id = ? AND warehouse = ?`,
		adj.Delta, time.Now().UTC(), adj.ProductID, warehouse); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *StockStore) ListByProduct(productID int64) ([]models.StockItem, error) {
	rows, err := s.db.Query(`
		SELECT id, product_id, warehouse, quantity, reorder_at, updated_at
		FROM stock_items WHERE product_id = ?`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.StockItem
	for rows.Next() {
		var si models.StockItem
		if err := rows.Scan(&si.ID, &si.ProductID, &si.Warehouse, &si.Quantity, &si.ReorderAt, &si.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, si)
	}
	return out, rows.Err()
}

// LowStock returns items at or below their reorder threshold — handy for a
// "needs restock" dashboard widget.
func (s *StockStore) LowStock() ([]models.StockItem, error) {
	rows, err := s.db.Query(`
		SELECT id, product_id, warehouse, quantity, reorder_at, updated_at
		FROM stock_items WHERE quantity <= reorder_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.StockItem
	for rows.Next() {
		var si models.StockItem
		if err := rows.Scan(&si.ID, &si.ProductID, &si.Warehouse, &si.Quantity, &si.ReorderAt, &si.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, si)
	}
	return out, rows.Err()
}
