package dto

import "smartcredit/internal/domain"

type Product struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Description   string  `json:"description,omitempty"`
	MinAmount     float64 `json:"minAmount"`
	MaxAmount     float64 `json:"maxAmount"`
	MinTermMonths int     `json:"minTermMonths"`
	MaxTermMonths int     `json:"maxTermMonths"`
	BaseRate      float64 `json:"baseRate"`
}

func FromProduct(p domain.Product) Product {
	return Product{p.Code, p.Name, p.Description, p.MinAmount, p.MaxAmount, p.MinTermMonths, p.MaxTermMonths, p.BaseRate}
}

func FromProducts(ps []domain.Product) []Product { return mapAll(ps, FromProduct) }

type LoanRequest struct {
	ProductCode string  `json:"productCode"`
	Purpose     string  `json:"purpose,omitempty"`
	Amount      float64 `json:"amount"`
	TermMonths  int     `json:"termMonths"`
}

type Calculation struct {
	Rate           float64 `json:"rate"`
	MonthlyPayment float64 `json:"monthlyPayment"`
	TotalPayment   float64 `json:"totalPayment"`
	Overpayment    float64 `json:"overpayment"`
}

func FromCalculation(c domain.LoanCalculation) Calculation {
	return Calculation{c.Rate, c.MonthlyPayment, c.TotalPayment, c.Overpayment}
}

type Recommendation struct {
	ProductCode    string  `json:"productCode"`
	MonthlyPayment float64 `json:"monthlyPayment"`
	Explanation    string  `json:"explanation"`
	Source         string  `json:"source"`
}

func FromRecommendation(r domain.Recommendation) Recommendation {
	return Recommendation{r.ProductCode, r.MonthlyPayment, r.Explanation, r.Source}
}
