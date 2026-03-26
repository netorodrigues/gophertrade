package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyOrder    = errors.New("order must have at least one item")
	ErrInvalidQuantity = errors.New("item quantity must be greater than zero")
	ErrInvalidPrice    = errors.New("item price must be non-negative")
)

type OrderStatus string

const (
	StatusCreated   OrderStatus = "CREATED"
	StatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID         uuid.UUID
	CreatedAt  time.Time
	TotalPrice int64
	Status     OrderStatus
	Items      []OrderItem
}

type OrderItem struct {
	OrderID   uuid.UUID
	ProductID uuid.UUID
	Quantity  int32
	UnitPrice int64
}

type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
}

func NewOrder(items []OrderItem) (*Order, error) {
	if len(items) == 0 {
		return nil, ErrEmptyOrder
	}

	var totalPrice int64
	for _, item := range items {
		if item.Quantity <= 0 {
			return nil, ErrInvalidQuantity
		}
		if item.UnitPrice < 0 {
			return nil, ErrInvalidPrice
		}
		totalPrice += int64(item.Quantity) * item.UnitPrice
	}

	orderID := uuid.New()
	for i := range items {
		items[i].OrderID = orderID
	}

	return &Order{
		ID:         orderID,
		CreatedAt:  time.Now(),
		TotalPrice: totalPrice,
		Status:     StatusCreated,
		Items:      items,
	}, nil
}
