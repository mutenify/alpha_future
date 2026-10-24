package scoring

import "smartcredit/internal/domain"

// PGap — вероятность, что месячный свободный поток X ~ N(μ, σ) с равномерно
// распределённой на срок подушкой C₀ не покроет D_old + A.
func PGap(mu, sigma, debtOld, payment, cashBalance float64, termMonths int) float64 {
	cushion := 0.0
	if termMonths > 0 {
		cushion = cashBalance / float64(termMonths)
	}
	threshold := debtOld + payment - cushion
	if sigma == 0 {
		if mu >= threshold {
			return 0
		}
		return 1
	}
	return Phi((threshold - mu) / sigma)
}

// MaxPayment — наибольший платёж A, при котором P_gap ≤ q и (D_old + A) ≤ k·μ.
func MaxPayment(p domain.FinancialProfile, pr domain.Product, termMonths int) float64 {
	cushion := 0.0
	if termMonths > 0 {
		cushion = p.CashBalance / float64(termMonths)
	}
	byGap := p.CashFlowMu + cushion - p.CurrentDebtPayments + p.CashFlowSigma*PhiInv(pr.GapThresholdQ)
	byShare := pr.PaymentShareK*p.CashFlowMu - p.CurrentDebtPayments
	return max(0, min(byGap, byShare))
}

// CashGapMetrics — блок «Слой 2» в карточке инспектора.
func CashGapMetrics(p domain.FinancialProfile, pr domain.Product, payment float64, termMonths int) domain.CashGap {
	total := p.CurrentDebtPayments + payment
	g := domain.CashGap{
		RequestedPayment: Round2(payment),
		PGap:             PGap(p.CashFlowMu, p.CashFlowSigma, p.CurrentDebtPayments, payment, p.CashBalance, termMonths),
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
