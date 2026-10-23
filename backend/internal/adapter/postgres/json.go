package postgres

import "smartcredit/internal/domain"

// JSON-формы jsonb-колонок: ключи как в комментариях миграций.

type monthJSON struct {
	Month  string  `json:"month"`
	Amount float64 `json:"amount"`
}

type factorJSON struct {
	Code         string  `json:"code"`
	Value        string  `json:"value"`
	Points       int     `json:"points"`
	Contribution float64 `json:"contribution"`
	Explanation  string  `json:"explanation,omitempty"`
}

type metricJSON struct {
	Code  string  `json:"code"`
	Value float64 `json:"value"`
	Color string  `json:"color"`
}

func mapSlice[A, B any](in []A, f func(A) B) []B {
	out := make([]B, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

func toMonthJSON(m domain.MonthAmount) monthJSON { return monthJSON{m.Month, m.Amount} }
func fromMonthJSON(m monthJSON) domain.MonthAmount {
	return domain.MonthAmount{Month: m.Month, Amount: m.Amount}
}

func toFactorJSON(f domain.Factor) factorJSON {
	return factorJSON{f.Code, f.Value, f.Points, f.Contribution, f.Explanation}
}

func fromFactorJSON(f factorJSON) domain.Factor {
	return domain.Factor{Code: f.Code, Value: f.Value, Points: f.Points, Contribution: f.Contribution, Explanation: f.Explanation}
}

func toMetricJSON(m domain.Metric) metricJSON { return metricJSON{m.Code, m.Value, m.Color} }
func fromMetricJSON(m metricJSON) domain.Metric {
	return domain.Metric{Code: m.Code, Value: m.Value, Color: m.Color}
}
