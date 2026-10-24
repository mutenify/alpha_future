package catalog

import (
	"context"
	"fmt"

	"smartcredit/internal/domain"
	"smartcredit/internal/scoring"
)

// Recommend — LLM-подбор с обязательной проверкой продуктовых ограничений и fallback на правила.
func (s *Service) Recommend(ctx context.Context, purpose string, amount float64, term int) (domain.Recommendation, error) {
	products, err := s.products.ListActive(ctx)
	if err != nil {
		return domain.Recommendation{}, err
	}
	if code, explanation, source, callErr := s.recommender.Recommend(ctx, purpose, amount, term); callErr == nil {
		for _, p := range products {
			if p.Code == code && p.Validate(amount, term) == nil {
				return domain.Recommendation{
					ProductCode: p.Code, MonthlyPayment: scoring.Round2(scoring.Payment(amount, term, p.BaseRate)),
					Explanation: explanation, Source: source,
				}, nil
			}
		}
	}
	for _, p := range products {
		if p.Validate(amount, term) == nil {
			return domain.Recommendation{
				ProductCode:    p.Code,
				MonthlyPayment: scoring.Round2(scoring.Payment(amount, term, p.BaseRate)),
				Explanation:    fmt.Sprintf("%s подходит под сумму и срок: %s", p.Name, p.Description),
				Source:         "RULES",
			}, nil
		}
	}
	return domain.Recommendation{}, domain.Invalid("нет продукта под эту сумму и срок")
}
