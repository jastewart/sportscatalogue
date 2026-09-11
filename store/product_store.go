package store

import (
	"database/sql"
	"time"

	"github.com/caleb/sports-catalog/models"
)

type ProductStore struct {
	db *sql.DB
}

func NewProductStore(db *sql.DB) *ProductStore {
	return &ProductStore{db: db}
}

func (s *ProductStore) Create(p *models.Product) (int64, error) {
	now := time.Now().UTC()
	res, err := s.db.Exec(`
		INSERT INTO products (sku, name, description, category_type, subcategory,
			price_cents, currency, image_url, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.SKU, p.Name, p.Description, string(p.CategoryType), p.Subcategory,
		p.PriceCents, p.Currency, p.ImageURL, p.Active, now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// List returns products, optionally filtered by category type — pass "" to
// get the full catalogue, or models.CategorySportsEquipment to get just
// sports equipment stock, etc.
func (s *ProductStore) List(categoryType models.CategoryType) ([]models.ProductWithStock, error) {
	query := `
		SELECT p.id, p.sku, p.name, p.description, p.category_type, p.subcategory,
			p.price_cents, p.currency, p.image_url, p.active, p.created_at, p.updated_at,
			COALESCE(SUM(si.quantity), 0) AS total_quantity
		FROM products p
		LEFT JOIN stock_items si ON si.product_id = p.id
	`
	args := []any{}
	if categoryType != "" {
		query += " WHERE p.category_type = ?"
		args = append(args, string(categoryType))
	}
	query += " GROUP BY p.id ORDER BY p.name ASC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.ProductWithStock
	for rows.Next() {
		var pw models.ProductWithStock
		var ct string
		if err := rows.Scan(&pw.ID, &pw.SKU, &pw.Name, &pw.Description, &ct, &pw.Subcategory,
			&pw.PriceCents, &pw.Currency, &pw.ImageURL, &pw.Active, &pw.CreatedAt, &pw.UpdatedAt,
			&pw.TotalQuantity); err != nil {
			return nil, err
		}
		pw.CategoryType = models.CategoryType(ct)
		out = append(out, pw)
	}
	return out, rows.Err()
}

func (s *ProductStore) GetByID(id int64) (*models.Product, error) {
	var p models.Product
	var ct string
	err := s.db.QueryRow(`
		SELECT id, sku, name, description, category_type, subcategory,
			price_cents, currency, image_url, active, created_at, updated_at
		FROM products WHERE id = ?`, id,
	).Scan(&p.ID, &p.SKU, &p.Name, &p.Description, &ct, &p.Subcategory,
		&p.PriceCents, &p.Currency, &p.ImageURL, &p.Active, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	p.CategoryType = models.CategoryType(ct)
	return &p, nil
}
