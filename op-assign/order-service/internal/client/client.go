package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
	"fmt"
)

type PaymentClient interface {
	Pay(ctx context.Context, orderID string, amount int64) (string, error)
}

type paymentClient struct {
	baseURL string
	client  *http.Client
}

func NewPaymentClient(url string) PaymentClient {
	return &paymentClient{
		baseURL: url,
		client: &http.Client{
			Timeout: 2 * time.Second,
		},
	}
}

func (c *paymentClient) Pay(ctx context.Context, orderID string, amount int64) (string, error) {
	body := map[string]interface{}{
		"order_id": orderID,
		"amount":   amount,
	}

	b, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/payments", bytes.NewBuffer(b))
	if err != nil {
    	return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res struct {
		Status string `json:"status"`
	}

	if resp.StatusCode != http.StatusOK {
    	return "", fmt.Errorf("payment service error: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	return res.Status, nil
}
