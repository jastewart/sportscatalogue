package main

import (
	"log"
	"net/http"

	"github.com/caleb/sports-catalog/config"
	"github.com/caleb/sports-catalog/handlers"
	"github.com/caleb/sports-catalog/payment"
	"github.com/caleb/sports-catalog/store"
)

func main() {
	cfg := config.Load()

	db, err := store.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	productStore := store.NewProductStore(db)
	stockStore := store.NewStockStore(db)
	orderStore := store.NewOrderStore(db)

	stripeClient := payment.NewStripeClient(
		cfg.StripeSecretKey,
		cfg.CheckoutSuccess,
		cfg.CheckoutCancelURL,
		cfg.StripeWebhookKey,
	)

	productHandler := handlers.NewProductHandler(productStore)
	stockHandler := handlers.NewStockHandler(stockStore)
	checkoutHandler := handlers.NewCheckoutHandler(productStore, stockStore, orderStore, stripeClient)

	mux := http.NewServeMux()

	// Catalogue — filter with ?category=sports_equipment|apparel|accessory|general
	mux.HandleFunc("GET /api/products", productHandler.ListProducts)
	mux.HandleFunc("POST /api/products", productHandler.CreateProduct)

	// Stock / inventory — separate from catalogue metadata
	mux.HandleFunc("POST /api/stock/adjust", stockHandler.AdjustStock)
	mux.HandleFunc("GET /api/stock/low", stockHandler.LowStock)

	// Payments (mock Stripe checkout)
	mux.HandleFunc("POST /api/checkout", checkoutHandler.CreateCheckout)
	mux.HandleFunc("POST /api/webhooks/stripe", checkoutHandler.StripeWebhook)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
