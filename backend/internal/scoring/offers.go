package scoring

import "smartcredit/internal/domain"

// Terms — сумма, срок и платёж одного варианта.
type Terms struct {
	Amount         float64
	TermMonths     int
	MonthlyPayment float64
	PGap           float64
}

// PointA — максимальная сумма при запрошенном сроке.
func PointA(p domain.FinancialProfile, pr domain.Product, term int, rate float64) (Terms, bool) {
	amount := min(Principal(MaxPayment(p, pr), term, rate), pr.MaxAmount)
	if amount < pr.MinAmount {
		return Terms{}, false
	}
	return terms(p, amount, term, rate), true
}

// PointB — запрошенная сумма на минимальном сроке, где платёж проходит лимит.
func PointB(p domain.FinancialProfile, pr domain.Product, amount float64, fromTerm int, rate float64) (Terms, bool) {
	limit := MaxPayment(p, pr)
	for t := fromTerm + 1; t <= pr.MaxTermMonths; t++ {
		if Payment(amount, t, rate) <= limit {
			return terms(p, amount, t, rate), true
		}
	}
	return Terms{}, false
}

// Fits — проходит ли запрошенная сумма лимит по потоку (S ≤ S*).
func Fits(p domain.FinancialProfile, pr domain.Product, amount float64, term int, rate float64) bool {
	return Payment(amount, term, rate) <= MaxPayment(p, pr)
}

func terms(p domain.FinancialProfile, amount float64, term int, rate float64) Terms {
	a := Payment(amount, term, rate)
	return Terms{
		Amount:         Round2(amount),
		TermMonths:     term,
		MonthlyPayment: Round2(a),
		PGap:           PGap(p.CashFlowMu, p.CashFlowSigma, p.CurrentDebtPayments, a),
	}
}
