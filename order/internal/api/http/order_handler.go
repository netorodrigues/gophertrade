package http

import (
	"encoding/json"
	"net/http"

	"gophertrade/order/internal/application"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OrderHandler struct {
	service *application.OrderService
}

func NewOrderHandler(service *application.OrderService) *OrderHandler {
	return &OrderHandler{
		service: service,
	}
}

func (h *OrderHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/health", HealthCheck)
	r.Post("/", h.CreateOrder)
	r.Get("/{id}", h.GetOrder)
	return r
}

type CreateOrderRequest struct {
	Items []struct {
		ProductID string `json:"product_id"`
		Quantity  int32  `json:"quantity"`
	} `json:"items"`
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	appReq := application.CreateOrderRequest{}
	for _, item := range req.Items {
		appReq.Items = append(appReq.Items, struct {
			ProductID string
			Quantity  int32
		}{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	order, err := h.service.CreateOrder(r.Context(), appReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid order ID", http.StatusBadRequest)
		return
	}

	order, err := h.service.GetOrder(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
