package config

import "os"

type Config struct {
	Port              string
	DBPath            string
	StripeSecretKey   string
	StripeWebhookKey  string
	CheckoutSuccess   string
	CheckoutCancelURL string
}

func Load() Config {
	return Config{
		Port:              getEnv("PORT", "8080"),
		DBPath:            getEnv("DB_PATH", "./catalog.db"),
		StripeSecretKey:   getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookKey:  getEnv("STRIPE_WEBHOOK_SECRET", ""),
		CheckoutSuccess:   getEnv("CHECKOUT_SUCCESS_URL", "http://localhost:8080/success"),
		CheckoutCancelURL: getEnv("CHECKOUT_CANCEL_URL", "http://localhost:8080/cancel"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
