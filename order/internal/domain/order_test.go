package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewOrder(t *testing.T) {
	productID1 := uuid.New()
	productID2 := uuid.New()

	tests := []struct {
		name          string
		items         []OrderItem
		expectedTotal int64
		expectedError error
	}{
		{
			name: "valid order with multiple items",
			items: []OrderItem{
				{ProductID: productID1, Quantity: 2, UnitPrice: 1000}, // 2000
				{ProductID: productID2, Quantity: 1, UnitPrice: 500},  // 500
			},
			expectedTotal: 2500,
			expectedError: nil,
		},
		{
			name:          "empty order",
			items:         []OrderItem{},
			expectedTotal: 0,
			expectedError: ErrEmptyOrder,
		},
		{
			name: "invalid quantity",
			items: []OrderItem{
				{ProductID: productID1, Quantity: 0, UnitPrice: 1000},
			},
			expectedTotal: 0,
			expectedError: ErrInvalidQuantity,
		},
		{
			name: "invalid price",
			items: []OrderItem{
				{ProductID: productID1, Quantity: 1, UnitPrice: -100},
			},
			expectedTotal: 0,
			expectedError: ErrInvalidPrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order, err := NewOrder(tt.items)
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, order)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, order)
				assert.Equal(t, tt.expectedTotal, order.TotalPrice)
				assert.Equal(t, StatusCreated, order.Status)
				assert.Len(t, order.Items, len(tt.items))
				for _, item := range order.Items {
					assert.Equal(t, order.ID, item.OrderID)
				}
			}
		})
	}
}
