package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/caleb/sports-catalog/models"
	"github.com/caleb/sports-catalog/store"
)

type ProductHandler struct {
	products *store.ProductStore
}

func NewProductHandler(products *store.ProductStore) *ProductHandler {
	return &ProductHandler{products: products}
}

// ListProducts handles GET /api/products and GET /api/products?category=sports_equipment
// The category filter is how callers separate "just sports equipment stock"
// from "the whole catalogue".
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	category := models.CategoryType(r.URL.Query().Get("category"))
	if category != "" && !category.Valid() {
		writeError(w, http.StatusBadRequest, "invalid category")
		return
	}

	items, err := h.products.List(category)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list products")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if p.SKU == "" || p.Name == "" || !p.CategoryType.Valid() {
		writeError(w, http.StatusBadRequest, "sku, name, and a valid category_type are required")
		return
	}
	if p.Currency == "" {
		p.Currency = "usd"
	}
	p.Active = true

	id, err := h.products.Create(&p)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create product")
		return
	}
	p.ID = id
	writeJSON(w, http.StatusCreated, p)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
