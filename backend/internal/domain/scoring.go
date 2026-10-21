package domain

// FinancialProfile — нормализованные входы скоринга на момент рассмотрения.
type FinancialProfile struct {
	ID                  string
	ApplicationID       string
	MonthlyRevenue      []MonthAmount
	RevenueCV           float64
	RevenueTrend        float64
	CashBalance         float64 // C₀
	CurrentDebtPayments float64 // D_old
	TaxDebt             float64
	Overdues12m         int
	MaxOverdueDays      int
	Inquiries3m         int
	BusinessAgeMonths   int
	OKVED               string
	IndustryCode        string
	IndustryMargin      float64
	CashFlowMu          float64 // μ
	CashFlowSigma       float64 // σ
	StopFactor          string  // LIQUIDATION | BANKRUPTCY | ACCOUNTS_BLOCKED
}

type MonthAmount struct {
	Month  string
	Amount float64
}

type Factor struct {
	Code         string
	Value        string
	Points       int
	Contribution float64
	Explanation  string
}

type ScoreResult struct {
	ID               string
	ApplicationID    string
	ProfileID        string
	ModelVersion     string
	Score            int
	PD               float64
	Factors          []Factor
	EquilibriumRate  float64
	InspectorMetrics []Metric
}

type Metric struct {
	Code  string
	Value float64
	Color string // GREEN | YELLOW | RED
}

type EnrichmentResult struct {
	Source  string // EGRUL | FNS | BKI
	Status  string // OK | ERROR | TIMEOUT
	Payload map[string]any
}

// ScoringDetails — всё для карточки инспектора.
type ScoringDetails struct {
	Result     ScoreResult
	Profile    FinancialProfile
	CashGap    *CashGap
	Enrichment []EnrichmentResult
}

type CashGap struct {
	RequestedPayment float64
	PGap             float64
	GapThresholdQ    float64
	DebtCoverage     float64
	SafetyMonths     float64
	Volatility       float64
}
