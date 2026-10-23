package mocks

import (
	"context"
	"time"

	"smartcredit/internal/usecase/port"
)

type egrulResponse struct {
	INN          string `json:"inn"`
	OGRN         string `json:"ogrn"`
	Name         string `json:"name"`
	LegalForm    string `json:"legalForm"`
	Status       string `json:"status"`
	RegisteredAt string `json:"registeredAt"`
	OKVED        string `json:"okved"`
	DirectorName string `json:"directorName"`
}

func (c *Client) Egrul(ctx context.Context, inn string) (port.EgrulData, error) {
	var r egrulResponse
	if err := c.do(ctx, "GET", "/egrul/"+inn, nil, &r); err != nil {
		return port.EgrulData{}, err
	}
	reg, _ := time.Parse(time.DateOnly, r.RegisteredAt)
	return port.EgrulData{
		INN: r.INN, OGRN: r.OGRN, Name: r.Name, LegalForm: r.LegalForm, Status: r.Status,
		OKVED: r.OKVED, DirectorName: r.DirectorName, RegisteredAt: reg,
	}, nil
}
