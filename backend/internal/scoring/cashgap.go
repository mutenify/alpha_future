package scoring

import "smartcredit/internal/domain"

// PGap — вероятность, что месячный свободный поток X ~ N(μ, σ) не покроет D_old + A.
// TODO(математик): сверить с моделью Мертона из документа (учёт подушки C₀).
func PGap(mu, sigma, debtOld, payment float64) float64 {
	if sigma == 0 {
		if mu >= debtOld+payment {
			return 0
		}
		return 1
	}
	return Phi((debtOld + payment - mu) / sigma)
}

// MaxPayment — наибольший платёж A, при котором P_gap ≤ q и (D_old + A) ≤ k·μ.
func MaxPayment(p domain.FinancialProfile, pr domain.Product) float64 {
	byGap := p.CashFlowMu - p.CurrentDebtPayments + p.CashFlowSigma*PhiInv(pr.GapThresholdQ)
	byShare := pr.PaymentShareK*p.CashFlowMu - p.CurrentDebtPayments
	return max(0, min(byGap, byShare))
}

// CashGapMetrics — блок «Слой 2» в карточке инспектора.
func CashGapMetrics(p domain.FinancialProfile, pr domain.Product, payment float64) domain.CashGap {
	total := p.CurrentDebtPayments + payment
	g := domain.CashGap{
		RequestedPayment: Round2(payment),
		PGap:             PGap(p.CashFlowMu, p.CashFlowSigma, p.CurrentDebtPayments, payment),
		GapThresholdQ:    pr.GapThresholdQ,
	}
	if total > 0 {
		g.DebtCoverage = p.CashFlowMu / total
		g.SafetyMonths = p.CashBalance / total
	}
	if p.CashFlowMu != 0 {
		g.Volatility = p.CashFlowSigma / p.CashFlowMu
	}
	return g
}
