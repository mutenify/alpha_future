package dto

import "smartcredit/internal/domain"

type FinancialProfile struct {
	MonthlyRevenue      []MonthAmount `json:"monthlyRevenue"`
	RevenueCV           float64       `json:"revenueCv"`
	CashBalance         float64       `json:"cashBalance"`
	CurrentDebtPayments float64       `json:"currentDebtPayments"`
	TaxDebt             float64       `json:"taxDebt"`
	BkiOverdues12m      int           `json:"bkiOverdues12m"`
	BkiInquiries3m      int           `json:"bkiInquiries3m"`
	BusinessAgeMonths   int           `json:"businessAgeMonths"`
	OKVED               string        `json:"okved,omitempty"`
	CashFlowMu          float64       `json:"cashFlowMu"`
	CashFlowSigma       float64       `json:"cashFlowSigma"`
}

type MonthAmount struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
}

func FromProfile(p domain.FinancialProfile) FinancialProfile {
	return FinancialProfile{
		MonthlyRevenue:      mapAll(p.MonthlyRevenue, func(m domain.MonthAmount) MonthAmount { return MonthAmount(m) }),
		RevenueCV:           p.RevenueCV,
		CashBalance:         p.CashBalance,
		CurrentDebtPayments: p.CurrentDebtPayments,
		TaxDebt:             p.TaxDebt,
		BkiOverdues12m:      p.Overdues12m,
		BkiInquiries3m:      p.Inquiries3m,
		BusinessAgeMonths:   p.BusinessAgeMonths,
		OKVED:               p.OKVED,
		CashFlowMu:          p.CashFlowMu,
		CashFlowSigma:       p.CashFlowSigma,
	}
}
