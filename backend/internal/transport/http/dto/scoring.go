package dto

import "smartcredit/internal/domain"

type ScoringDetails struct {
	PD               float64            `json:"pd"`
	Score            int                `json:"score"`
	ModelVersion     string             `json:"modelVersion"`
	EquilibriumRate  float64            `json:"equilibriumRate,omitempty"`
	Factors          []Factor           `json:"factors"`
	InspectorMetrics []Metric           `json:"inspectorMetrics,omitempty"`
	FinancialProfile FinancialProfile   `json:"financialProfile"`
	CashGap          *CashGap           `json:"cashGap,omitempty"`
	Enrichment       []EnrichmentSource `json:"enrichment"`
}

type Factor struct {
	Code         string  `json:"code"`
	Value        string  `json:"value,omitempty"`
	Contribution float64 `json:"contribution"`
	Explanation  string  `json:"explanation,omitempty"`
}

type Metric struct {
	Code  string  `json:"code"`
	Value float64 `json:"value"`
	Color string  `json:"color"`
}

type CashGap struct {
	RequestedPayment float64 `json:"requestedPayment"`
	PGap             float64 `json:"pGap"`
	GapThresholdQ    float64 `json:"gapThresholdQ"`
	DebtCoverage     float64 `json:"debtCoverage"`
	SafetyMonths     float64 `json:"safetyMonths"`
	Volatility       float64 `json:"volatility"`
}

type EnrichmentSource struct {
	Source  string         `json:"source"`
	Status  string         `json:"status"`
	Payload map[string]any `json:"payload,omitempty"`
}

func FromScoring(d domain.ScoringDetails) ScoringDetails {
	out := ScoringDetails{
		PD: d.Result.PD, Score: d.Result.Score, ModelVersion: d.Result.ModelVersion,
		EquilibriumRate: d.Result.EquilibriumRate,
		Factors: mapAll(d.Result.Factors, func(f domain.Factor) Factor {
			return Factor{f.Code, f.Value, f.Contribution, f.Explanation}
		}),
		InspectorMetrics: mapAll(d.Result.InspectorMetrics, func(m domain.Metric) Metric { return Metric(m) }),
		FinancialProfile: FromProfile(d.Profile),
		Enrichment: mapAll(d.Enrichment, func(e domain.EnrichmentResult) EnrichmentSource {
			return EnrichmentSource(e)
		}),
	}
	if g := d.CashGap; g != nil {
		cg := CashGap(*g)
		out.CashGap = &cg
	}
	return out
}
