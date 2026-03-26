package http

import (
	"encoding/json"
	"net/http"

	"gophertrade/order/internal/infrastructure/persistence/es"
	"gophertrade/order/internal/infrastructure/persistence/firestore"

	"github.com/go-chi/chi/v5"
)

type QueryHandler struct {
	fsRepo *firestore.OrderReadRepository
	esRepo *es.OrderSearchRepository
}

func NewQueryHandler(fsRepo *firestore.OrderReadRepository, esRepo *es.OrderSearchRepository) *QueryHandler {
	return &QueryHandler{
		fsRepo: fsRepo,
		esRepo: esRepo,
	}
}

func (h *QueryHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/search", h.SearchOrders)
	r.Get("/view/{id}", h.ViewOrder)
	return r
}

func (h *QueryHandler) SearchOrders(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	orders, err := h.esRepo.Search(r.Context(), q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *QueryHandler) ViewOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	order, err := h.fsRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
