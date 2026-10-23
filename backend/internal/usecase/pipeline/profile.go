package pipeline

import (
	"math"
	"time"

	"smartcredit/internal/domain"
)

// buildProfile нормализует ответы источников во входы скоринга.
// μ и σ свободного потока = выручка × маржа отрасли.
// TODO(математик): сверить расчёт μ, σ и тренда с generate_companies.py.
func buildProfile(app domain.Application, d sourceData, ind domain.Industry) domain.FinancialProfile {
	rev := make([]domain.MonthAmount, len(d.fns.MonthlyRevenue))
	amounts := make([]float64, len(rev))
	for i, m := range d.fns.MonthlyRevenue {
		rev[i] = domain.MonthAmount{Month: m.Month, Amount: m.Amount}
		amounts[i] = m.Amount
	}
	mean, std := meanStd(amounts)
	p := domain.FinancialProfile{
		ApplicationID:       app.ID,
		MonthlyRevenue:      rev,
		RevenueTrend:        trend(amounts),
		CashBalance:         totalBalance(d),
		CurrentDebtPayments: d.bki.MonthlyPayments,
		TaxDebt:             d.fns.TaxDebt,
		Overdues12m:         d.bki.Overdues12m,
		MaxOverdueDays:      d.bki.MaxOverdueDays,
		Inquiries3m:         d.bki.Inquiries3m,
		BusinessAgeMonths:   monthsSince(d.egrul.RegisteredAt),
		OKVED:               d.egrul.OKVED,
		IndustryCode:        ind.Code,
		IndustryMargin:      ind.Margin,
		CashFlowMu:          mean * ind.Margin,
		CashFlowSigma:       std * ind.Margin,
		StopFactor:          stopFactor(d),
	}
	if mean > 0 {
		p.RevenueCV = std / mean
	}
	return p
}

func stopFactor(d sourceData) string {
	switch {
	case d.egrul.Status == "LIQUIDATING" || d.egrul.Status == "LIQUIDATED":
		return "LIQUIDATION"
	case d.egrul.Status == "BANKRUPTCY":
		return "BANKRUPTCY"
	case d.fns.AccountsBlocked:
		return "ACCOUNTS_BLOCKED"
	}
	return ""
}

func totalBalance(d sourceData) float64 {
	var sum float64
	for _, a := range d.accounts {
		sum += a.Balance
	}
	return sum
}

func meanStd(xs []float64) (mean, std float64) {
	if len(xs) == 0 {
		return 0, 0
	}
	for _, x := range xs {
		mean += x
	}
	mean /= float64(len(xs))
	for _, x := range xs {
		std += (x - mean) * (x - mean)
	}
	return mean, math.Sqrt(std / float64(len(xs)))
}

// trend — относительный рост: последние 3 месяца к первым 3.
func trend(xs []float64) float64 {
	if len(xs) < 6 {
		return 0
	}
	first, _ := meanStd(xs[:3])
	last, _ := meanStd(xs[len(xs)-3:])
	if first == 0 {
		return 0
	}
	return last/first - 1
}

func monthsSince(t time.Time) int {
	now := time.Now()
	return (now.Year()-t.Year())*12 + int(now.Month()-t.Month())
}
