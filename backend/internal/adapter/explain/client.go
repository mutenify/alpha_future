package explain

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"smartcredit/internal/usecase/port"
)

type Client struct {
	url      string
	http     *http.Client
	fallback Template
}

func New(url string, timeout time.Duration) *Client {
	return &Client{url: url, http: &http.Client{Timeout: timeout}}
}

func (c *Client) Explain(ctx context.Context, in port.ExplainInput) (string, string) {
	body, err := json.Marshal(map[string]any{
		"outcome": in.Outcome, "score": in.Score, "pd": in.PD, "pGap": in.PGap, "factors": in.Factors,
	})
	if err != nil {
		return c.fallback.Explain(ctx, in)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/llm/explain", bytes.NewReader(body))
	if err != nil {
		return c.fallback.Explain(ctx, in)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return c.fallback.Explain(ctx, in)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return c.fallback.Explain(ctx, in)
	}
	var out struct {
		DecisionReason string `json:"decisionReason"`
		Source         string `json:"source"`
	}
	if json.NewDecoder(resp.Body).Decode(&out) != nil || out.DecisionReason == "" || out.Source != "LLM" {
		return c.fallback.Explain(ctx, in)
	}
	return out.DecisionReason, "LLM"
}
