# sports-catalog

A starter Go codebase for a product catalogue (sports equipment + general
products) with separate stock tracking and a mock Stripe checkout flow.

## Why this structure

- **`models/product.go` vs `models/stock.go`** — this is the "differentiation"
  you asked for. `Product` is catalogue metadata (name, price, description,
  `category_type`) and never stores a quantity. `StockItem` is a separate
  table keyed by `(product_id, warehouse)`, so a product can have stock in
  multiple warehouses, and restocking/fulfillment never touches catalogue rows.
  Filter the catalogue by `category_type=sports_equipment` vs `apparel`,
  `accessory`, `general` to separate sports gear from the rest of the shop.
- **`store/`** — thin SQL layer over `modernc.org/sqlite` (pure Go, no cgo —
  same rationale as using Dapper over raw ADO.NET: it keeps queries explicit
  and testable without an ORM's magic).
- **`payment/stripe_client.go`** — wraps Stripe Checkout Sessions. In test
  mode this is a fully working mock: use card `4242 4242 4242 4242`, any
  future expiry, any CVC. No real money moves.
- **`handlers/`** — plain `net/http` using Go 1.22's method-aware
  `ServeMux` patterns (`"GET /api/products"`), so no third-party router is
  required.

## Project layout

```
sports-catalog/
├── main.go                  # wiring + routes
├── config/config.go         # env-based config
├── models/
│   ├── product.go           # catalogue entry
│   ├── stock.go             # inventory, separate from catalogue
│   └── order.go             # checkout/order records
├── store/
│   ├── db.go                # sqlite connection + schema
│   ├── product_store.go
│   ├── stock_store.go
│   └── order_store.go
├── handlers/
│   ├── product_handler.go
│   ├── stock_handler.go
│   └── checkout_handler.go
└── payment/
    └── stripe_client.go      # Stripe Checkout Session + webhook verification
```

## Setup

```bash
cp .env.example .env
# fill in your Stripe TEST secret key and webhook secret
go mod tidy       # fetches stripe-go and modernc.org/sqlite
go run .
```

Forward Stripe webhook events to your local server with the Stripe CLI:

```bash
stripe listen --forward-to localhost:8080/api/webhooks/stripe
```

## API quick reference

| Method | Path                     | Purpose                                   |
|--------|--------------------------|--------------------------------------------|
| GET    | `/api/products`          | List catalogue (`?category=` to filter)   |
| POST   | `/api/products`          | Add a catalogue item                      |
| POST   | `/api/stock/adjust`      | Restock or deduct stock (`delta`)         |
| GET    | `/api/stock/low`         | Items at/below reorder threshold          |
| POST   | `/api/checkout`          | Start a mock Stripe checkout session      |
| POST   | `/api/webhooks/stripe`   | Stripe calls this on payment completion   |

Example: create a sports-equipment product

```bash
curl -X POST localhost:8080/api/products -d '{
  "sku": "BASK-001",
  "name": "Official Size Basketball",
  "category_type": "sports_equipment",
  "subcategory": "Basketball",
  "price_cents": 2999,
  "currency": "usd"
}'
```

Example: start checkout

```bash
curl -X POST localhost:8080/api/checkout -d '{
  "customer_email": "test@example.com",
  "items": [{"product_id": 1, "quantity": 2}]
}'
```

## Known extension points (left intentionally open)

- `CheckoutHandler.deductStockForSession` needs `OrderStore` extended to
  load `order_items` by session ID, then call `StockStore.Adjust` per line
  item with a negative delta once payment is confirmed.
- No auth/middleware layer yet — add an API-key or JWT check on the mutating
  routes before exposing this beyond localhost.
- `go.sum` isn't included since this container has no network access to the
  Go module proxy — run `go mod tidy` on your machine to generate it.
