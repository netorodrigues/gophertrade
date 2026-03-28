package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	httpapi "gophertrade/order/internal/api/http"
	"gophertrade/order/internal/application"
	"gophertrade/order/internal/domain"
	inventoryv1 "gophertrade/proto/inventory/v1"

	"github.com/google/uuid"
	"github.com/go-chi/chi/v5"
)

type mockOrderRepo struct {
	orders map[uuid.UUID]*domain.Order
}

func (m *mockOrderRepo) Create(ctx context.Context, order *domain.Order) error {
	m.orders[order.ID] = order
	return nil
}

func (m *mockOrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	if o, ok := m.orders[id]; ok {
		return o, nil
	}
	return nil, fmt.Errorf("order not found")
}

type mockAppInventoryService struct {
	price int64
}

func (m *mockAppInventoryService) GetProduct(ctx context.Context, productID string) (*inventoryv1.GetProductResponse, error) {
	return &inventoryv1.GetProductResponse{
		ProductId:  productID,
		PriceCents: m.price,
	}, nil
}

func (m *mockAppInventoryService) BatchUpdateStock(ctx context.Context, items map[string]int32) error {
	return nil
}

type mockOrderPublisher struct{}

func (m *mockOrderPublisher) PublishOrderCreated(ctx context.Context, order *domain.Order) error {
	return nil
}

func setupService() *application.OrderService {
	repo := &mockOrderRepo{orders: make(map[uuid.UUID]*domain.Order)}
	inv := &mockAppInventoryService{price: 1500}
	pub := &mockOrderPublisher{}
	return application.NewOrderService(repo, inv, pub)
}

func TestHTTP_CreateAndGetOrder(t *testing.T) {
	svc := setupService()
	handler := httpapi.NewOrderHandler(svc)
	router := chi.NewRouter()
	router.Post("/", handler.CreateOrder)

	// Create Order
	reqBody := `{"items":[{"product_id":"` + uuid.New().String() + `","quantity":2}]}`
	req, _ := http.NewRequest("POST", "/", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Fatalf("handler returned wrong status code: got %v want %v. body: %s", status, http.StatusCreated, rr.Body.String())
	}

	var createdOrder domain.Order
	if err := json.NewDecoder(rr.Body).Decode(&createdOrder); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if createdOrder.TotalPrice != 3000 {
		t.Errorf("expected total price 3000, got %d", createdOrder.TotalPrice)
	}
}