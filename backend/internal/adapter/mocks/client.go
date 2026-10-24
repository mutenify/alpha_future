// Package mocks — HTTP-клиент Mock Gateway (api/mocks-openapi.yaml), реализует port.Sources.
package mocks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"smartcredit/internal/domain"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{Timeout: timeout}}
}

func (c *Client) Health(ctx context.Context) error {
	var out struct {
		Status string `json:"status"`
	}
	if err := c.do(ctx, http.MethodGet, "/health", nil, &out); err != nil {
		return err
	}
	if out.Status != "ok" {
		return domain.ErrSourceUnavailable
	}
	return nil
}

func (c *Client) Recommend(ctx context.Context, purpose string, amount float64, termMonths int) (string, string, string, error) {
	var out struct {
		ProductCode string `json:"productCode"`
		Explanation string `json:"explanation"`
		Source      string `json:"source"`
	}
	err := c.do(ctx, http.MethodPost, "/llm/recommend", map[string]any{
		"purpose": purpose, "amount": amount, "termMonths": termMonths,
	}, &out)
	return out.ProductCode, out.Explanation, out.Source, err
}

// do: 404 → domain.ErrNotFound, таймаут и 5xx → domain.ErrSourceUnavailable (конвейер ретраит).
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %s: %v", domain.ErrSourceUnavailable, path, err)
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotFound:
		return domain.ErrNotFound
	case resp.StatusCode >= 300:
		return fmt.Errorf("%w: %s: HTTP %d", domain.ErrSourceUnavailable, path, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return errors.Join(domain.ErrSourceUnavailable, err)
	}
	return nil
}
