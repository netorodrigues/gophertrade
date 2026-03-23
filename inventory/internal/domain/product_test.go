package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProduct(t *testing.T) {
	name := "Gopher Plushie"
	price := int64(2500)
	initialStock := int64(100)

	product := NewProduct(name, price, initialStock)

	assert.NotNil(t, product.ID)
	assert.Equal(t, name, product.Name)
	assert.Equal(t, price, product.Price)
	assert.Equal(t, initialStock, product.StockQuantity)
	assert.Equal(t, int32(1), product.Version)
	assert.NotZero(t, product.CreatedAt)
	assert.NotZero(t, product.UpdatedAt)
}

func TestUpdateStock(t *testing.T) {
	product := NewProduct("Gopher Plushie", 2500, 100)

	t.Run("Increment stock", func(t *testing.T) {
		err := product.UpdateStock(10)
		assert.NoError(t, err)
		assert.Equal(t, int64(110), product.StockQuantity)
	})

	t.Run("Decrement stock", func(t *testing.T) {
		err := product.UpdateStock(-20)
		assert.NoError(t, err)
		assert.Equal(t, int64(90), product.StockQuantity)
	})

	t.Run("Insufficient stock", func(t *testing.T) {
		err := product.UpdateStock(-100)
		assert.ErrorIs(t, err, ErrInsufficientStock)
		assert.Equal(t, int64(90), product.StockQuantity)
	})
}
