package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // pure-Go sqlite driver, no cgo needed
)

const schema = `
CREATE TABLE IF NOT EXISTS products (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	sku           TEXT NOT NULL UNIQUE,
	name          TEXT NOT NULL,
	description   TEXT NOT NULL DEFAULT '',
	category_type TEXT NOT NULL,
	subcategory   TEXT NOT NULL DEFAULT '',
	price_cents   INTEGER NOT NULL,
	currency      TEXT NOT NULL DEFAULT 'usd',
	image_url     TEXT NOT NULL DEFAULT '',
	active        INTEGER NOT NULL DEFAULT 1,
	created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_products_category_type ON products(category_type);

CREATE TABLE IF NOT EXISTS stock_items (
	id          INTEGER PRIMARY KEY AUTOINCREMENT,
	product_id  INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
	warehouse   TEXT NOT NULL DEFAULT 'main',
	quantity    INTEGER NOT NULL DEFAULT 0,
	reorder_at  INTEGER NOT NULL DEFAULT 5,
	updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(product_id, warehouse)
);

CREATE TABLE IF NOT EXISTS orders (
	id                 INTEGER PRIMARY KEY AUTOINCREMENT,
	stripe_session_id  TEXT NOT NULL UNIQUE,
	status             TEXT NOT NULL DEFAULT 'pending',
	total_cents        INTEGER NOT NULL,
	currency           TEXT NOT NULL DEFAULT 'usd',
	customer_email     TEXT NOT NULL DEFAULT '',
	created_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS order_items (
	id               INTEGER PRIMARY KEY AUTOINCREMENT,
	order_id         INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
	product_id       INTEGER NOT NULL REFERENCES products(id),
	sku              TEXT NOT NULL,
	name             TEXT NOT NULL,
	quantity         INTEGER NOT NULL,
	unit_price_cents INTEGER NOT NULL
);
`

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	// SQLite handles a single writer well; keep this modest so we don't
	// hit "database is locked" under concurrent requests.
	db.SetMaxOpenConns(1)
	return db, nil
}
