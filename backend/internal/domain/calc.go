package domain

type LoanCalculation struct {
	Rate           float64
	MonthlyPayment float64
	TotalPayment   float64
	Overpayment    float64
}

type Recommendation struct {
	ProductCode    string
	MonthlyPayment float64
	Explanation    string
	Source         string // LLM | RULES
}
