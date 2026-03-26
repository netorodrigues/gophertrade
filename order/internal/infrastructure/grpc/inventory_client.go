package grpc

import (
	"context"
	"fmt"

	inventoryv1 "gophertrade/proto/inventory/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InventoryClient struct {
	client inventoryv1.InventoryServiceClient
}

func NewInventoryClient(addr string) (*InventoryClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to inventory service: %w", err)
	}

	return &InventoryClient{
		client: inventoryv1.NewInventoryServiceClient(conn),
	}, nil
}

func (c *InventoryClient) BatchUpdateStock(ctx context.Context, items map[string]int32) error {
	var updates []*inventoryv1.StockUpdateItem
	for productID, delta := range items {
		updates = append(updates, &inventoryv1.StockUpdateItem{
			ProductId:    productID,
			QuantityDelta: delta,
		})
	}

	resp, err := c.client.BatchUpdateStock(ctx, &inventoryv1.BatchUpdateStockRequest{
		Updates: updates,
	})
	if err != nil {
		return fmt.Errorf("gRPC batch update stock failed: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("inventory update failed: %s", resp.Message)
	}

	return nil
}

func (c *InventoryClient) GetProduct(ctx context.Context, productID string) (*inventoryv1.GetProductResponse, error) {
	return c.client.GetProduct(ctx, &inventoryv1.GetProductRequest{
		ProductId: productID,
	})
}
