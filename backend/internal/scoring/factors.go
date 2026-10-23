package scoring

import "smartcredit/internal/domain"

// factorValues — значения факторов карты из профиля. Ключи совпадают с scorecard_bins.factor.
func factorValues(p domain.FinancialProfile) map[string]float64 {
	debtLoad := 0.0
	if p.CashFlowMu > 0 {
		debtLoad = p.CurrentDebtPayments / p.CashFlowMu // TODO(математик): знаменатель μ или выручка
	}
	return map[string]float64{
		"age_months":       float64(p.BusinessAgeMonths),
		"overdues12m":      float64(p.Overdues12m),
		"max_overdue_days": float64(p.MaxOverdueDays),
		"inquiries3m":      float64(p.Inquiries3m),
		"tax_debt":         p.TaxDebt,
		"revenue_trend":    p.RevenueTrend,
		"revenue_cv":       p.RevenueCV,
		"debt_load":        debtLoad,
	}
}
