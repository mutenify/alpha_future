package company

import (
	"context"

	"smartcredit/internal/domain"
	"smartcredit/internal/usecase/port"
)

type Service struct{ sources port.Sources }

func New(s port.Sources) *Service { return &Service{sources: s} }

// Lookup — «Заполнить по ИНН» (мок ЕГРЮЛ).
func (s *Service) Lookup(ctx context.Context, inn string) (domain.Company, error) {
	e, err := s.sources.Egrul(ctx, inn)
	if err != nil {
		return domain.Company{}, err
	}
	return FromEgrul(e), nil
}

func FromEgrul(e port.EgrulData) domain.Company {
	reg := e.RegisteredAt
	return domain.Company{
		INN: e.INN, OGRN: e.OGRN, Name: e.Name,
		LegalForm: domain.LegalForm(e.LegalForm), OKVED: e.OKVED, RegisteredAt: &reg,
	}
}
