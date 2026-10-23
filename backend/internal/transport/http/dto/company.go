package dto

import "smartcredit/internal/domain"

type Company struct {
	INN          string           `json:"inn"`
	OGRN         string           `json:"ogrn,omitempty"`
	Name         string           `json:"name"`
	LegalForm    domain.LegalForm `json:"legalForm"`
	OKVED        string           `json:"okved,omitempty"`
	RegisteredAt string           `json:"registeredAt,omitempty"`
}

func FromCompany(c domain.Company) Company {
	out := Company{INN: c.INN, OGRN: c.OGRN, Name: c.Name, LegalForm: c.LegalForm, OKVED: c.OKVED}
	if c.RegisteredAt != nil {
		out.RegisteredAt = c.RegisteredAt.Format("2006-01-02")
	}
	return out
}
