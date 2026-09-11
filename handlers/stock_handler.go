package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/caleb/sports-catalog/models"
	"github.com/caleb/sports-catalog/store"
)

type StockHandler struct {
	stock *store.StockStore
}

func NewStockHandler(stock *store.StockStore) *StockHandler {
	return &StockHandler{stock: stock}
}

// AdjustStock handles POST /api/stock/adjust — used both for manual restocks
// (positive delta) and internally when an order is paid (negative delta).
func (h *StockHandler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	var adj models.StockAdjustment
	if err := json.NewDecoder(r.Body).Decode(&adj); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if adj.ProductID == 0 || adj.Delta == 0 {
		writeError(w, http.StatusBadRequest, "product_id and a non-zero delta are required")
		return
	}

	if err := h.stock.Adjust(adj); err != nil {
		if errors.Is(err, store.ErrInsufficientStock) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to adjust stock")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// LowStock handles GET /api/stock/low — for a restock dashboard.
func (h *StockHandler) LowStock(w http.ResponseWriter, r *http.Request) {
	items, err := h.stock.LowStock()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch low stock")
		return
	}
	writeJSON(w, http.StatusOK, items)
}
