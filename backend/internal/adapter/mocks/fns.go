package mocks

import (
	"context"

	"smartcredit/internal/usecase/port"
)

type fnsResponse struct {
	MonthlyRevenue []struct {
		Month  string  `json:"month"`
		Amount float64 `json:"amount"`
	} `json:"monthlyRevenue"`
	TaxDebt         float64 `json:"taxDebt"`
	AccountsBlocked bool    `json:"accountsBlocked"`
}

func (c *Client) Fns(ctx context.Context, inn string) (port.FnsData, error) {
	var r fnsResponse
	if err := c.do(ctx, "GET", "/fns/"+inn, nil, &r); err != nil {
		return port.FnsData{}, err
	}
	d := port.FnsData{TaxDebt: r.TaxDebt, AccountsBlocked: r.AccountsBlocked}
	for _, m := range r.MonthlyRevenue {
		d.MonthlyRevenue = append(d.MonthlyRevenue, port.MonthAmount{Month: m.Month, Amount: m.Amount})
	}
	return d, nil
}
