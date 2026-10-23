package mocks

import (
	"context"
	"time"

	"smartcredit/internal/usecase/port"
)

type accountsResponse struct {
	Accounts []struct {
		Number  string  `json:"number"`
		Balance float64 `json:"balance"`
	} `json:"accounts"`
}

func (c *Client) Accounts(ctx context.Context, inn string) ([]port.AbsAccount, error) {
	var r accountsResponse
	if err := c.do(ctx, "GET", "/abs/"+inn+"/accounts", nil, &r); err != nil {
		return nil, err
	}
	out := make([]port.AbsAccount, len(r.Accounts))
	for i, a := range r.Accounts {
		out[i] = port.AbsAccount{Number: a.Number, Balance: a.Balance}
	}
	return out, nil
}

type disburseRequest struct {
	ContractNumber string  `json:"contractNumber"`
	AccountNumber  string  `json:"accountNumber"`
	Amount         float64 `json:"amount"`
}

func (c *Client) Disburse(ctx context.Context, contractNumber, account string, amount float64) (time.Time, error) {
	var r struct {
		DisbursedAt time.Time `json:"disbursedAt"`
	}
	err := c.do(ctx, "POST", "/abs/disbursements", disburseRequest{contractNumber, account, amount}, &r)
	return r.DisbursedAt, err
}
