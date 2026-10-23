package catalog

import (
	"context"

	"smartcredit/internal/domain"
	"smartcredit/internal/scoring"
)

// Calculate — калькулятор «Подобрать кредит» по базовой ставке, без скоринга.
func (s *Service) Calculate(ctx context.Context, code string, amount float64, term int) (domain.LoanCalculation, error) {
	p, err := s.products.ByCode(ctx, code)
	if err != nil {
		return domain.LoanCalculation{}, err
	}
	if err := p.Validate(amount, term); err != nil {
		return domain.LoanCalculation{}, err
	}
	a := scoring.Payment(amount, term, p.BaseRate)
	total := a * float64(term)
	return domain.LoanCalculation{
		Rate:           p.BaseRate,
		MonthlyPayment: scoring.Round2(a),
		TotalPayment:   scoring.Round2(total),
		Overpayment:    scoring.Round2(total - amount),
	}, nil
}
