package rules

import (
	"context"

	"smartcredit/internal/domain"
	"smartcredit/internal/usecase/port"
)

type Service struct {
	products port.ProductRepository
	rules    port.RulesRepository
}

func New(p port.ProductRepository, r port.RulesRepository) *Service {
	return &Service{products: p, rules: r}
}

// Get — вкладка «Продукты и правила».
func (s *Service) Get(ctx context.Context) (domain.Rules, error) {
	var (
		r   domain.Rules
		err error
	)
	if r.Products, err = s.products.ListActive(ctx); err != nil {
		return r, err
	}
	if r.Scorecard, err = s.rules.ActiveScorecard(ctx); err != nil {
		return r, err
	}
	r.Industries, err = s.rules.Industries(ctx)
	return r, err
}
