package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	inventoryhttp "gophertrade/inventory/internal/api/http"
	"gophertrade/inventory/internal/application"
	"gophertrade/inventory/internal/domain"
	"gophertrade/inventory/internal/infrastructure/persistence/postgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProductHTTPHandler(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewProductRepository(pool)
	service := application.NewProductService(repo, nil) // No publisher for this test
	handler := inventoryhttp.NewProductHandler(service)
	router := handler.Routes()

	ctx := context.Background()

	var productID string

	t.Run("POST / creates a product", func(t *testing.T) {
		reqBody := inventoryhttp.CreateProductRequest{
			Name:         "Gopher Plushie",
			PriceCents:   2500,
			InitialStock: 100,
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(ctx, "POST", "/", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		
		var product domain.Product
		err := json.Unmarshal(rr.Body.Bytes(), &product)
		require.NoError(t, err)
		assert.Equal(t, reqBody.Name, product.Name)
		productID = product.ID.String()
	})

	t.Run("GET /{id} returns a product", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(ctx, "GET", "/"+productID, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var product domain.Product
		err := json.Unmarshal(rr.Body.Bytes(), &product)
		require.NoError(t, err)
		assert.Equal(t, "Gopher Plushie", product.Name)
	})

	t.Run("POST /{id}/stock updates stock", func(t *testing.T) {
		reqBody := inventoryhttp.UpdateStockRequest{Delta: -10}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequestWithContext(ctx, "POST", "/"+productID+"/stock", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)

		// Verify stock
		p, err := repo.GetByID(ctx, uuid.MustParse(productID))
		assert.NoError(t, err)
		assert.Equal(t, int64(90), p.StockQuantity)
	})
}
