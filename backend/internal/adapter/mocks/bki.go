package mocks

import (
	"context"

	"smartcredit/internal/usecase/port"
)

type bkiResponse struct {
	Overdues12m    int `json:"overdues12m"`
	MaxOverdueDays int `json:"maxOverdueDays"`
	Inquiries3m    int `json:"inquiries3m"`
	Obligations    []struct {
		MonthlyPayment float64 `json:"monthlyPayment"`
	} `json:"obligations"`
}

func (c *Client) Bki(ctx context.Context, inn string) (port.BkiData, error) {
	var r bkiResponse
	if err := c.do(ctx, "GET", "/bki/"+inn, nil, &r); err != nil {
		return port.BkiData{}, err
	}
	d := port.BkiData{Overdues12m: r.Overdues12m, MaxOverdueDays: r.MaxOverdueDays, Inquiries3m: r.Inquiries3m}
	for _, o := range r.Obligations {
		d.MonthlyPayments += o.MonthlyPayment
	}
	return d, nil
}
