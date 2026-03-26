package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8"
)

type OrderSearchRepository struct {
	client *elasticsearch.Client
}

func NewOrderSearchRepository(client *elasticsearch.Client) *OrderSearchRepository {
	return &OrderSearchRepository{client: client}
}

type OrderDoc struct {
	OrderID    string `json:"order_id"`
	TotalPrice int64  `json:"total_price"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

func (r *OrderSearchRepository) Save(ctx context.Context, order *OrderDoc) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order doc: %w", err)
	}

	resp, err := r.client.Index(
		"orders",
		bytes.NewReader(data),
		r.client.Index.WithDocumentID(order.OrderID),
		r.client.Index.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to index order: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("elasticsearch error: %s", body)
	}

	return nil
}

func (r *OrderSearchRepository) Search(ctx context.Context, query string) ([]OrderDoc, error) {
	var buf bytes.Buffer
	q := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"status": query,
			},
		},
	}
	if err := json.NewEncoder(&buf).Encode(q); err != nil {
		return nil, fmt.Errorf("failed to encode query: %w", err)
	}

	resp, err := r.client.Search(
		r.client.Search.WithContext(ctx),
		r.client.Search.WithIndex("orders"),
		r.client.Search.WithBody(&buf),
		r.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search orders: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("elasticsearch error: %s", body)
	}

	var res struct {
		Hits struct {
			Hits []struct {
				Source OrderDoc `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %w", err)
	}

	orders := make([]OrderDoc, 0, len(res.Hits.Hits))
	for _, hit := range res.Hits.Hits {
		orders = append(orders, hit.Source)
	}

	return orders, nil
}
