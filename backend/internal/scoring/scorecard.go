package scoring

import (
	"fmt"
	"math"

	"smartcredit/internal/domain"
)

const categoricalFactor = "okved_group"

// PD = 1 / (1 + base_odds · 2^((score − base_score) / pdo)).
func PD(score int, m domain.Scorecard) float64 {
	return 1 / (1 + m.BaseOdds*math.Pow(2, float64(score-m.BaseScore)/float64(m.PDO)))
}

// Score — сумма баллов по факторам и PD. Contribution фактора — насколько PD выше,
// чем был бы при лучшем бине этого фактора.
func Score(p domain.FinancialProfile, m domain.Scorecard) (int, float64, []domain.Factor) {
	values := factorValues(p)
	best := bestPoints(m.Bins)
	var (
		score   int
		matched []domain.Bin
	)
	for _, b := range m.Bins {
		if ok := binMatches(b, values, p.IndustryCode); ok {
			score += b.Points
			matched = append(matched, b)
		}
	}
	pd := PD(score, m)
	factors := make([]domain.Factor, 0, len(matched))
	for _, b := range matched {
		factors = append(factors, domain.Factor{
			Code:         b.Factor,
			Value:        valueLabel(b.Factor, values, p.IndustryCode),
			Points:       b.Points,
			Contribution: pd - PD(score+best[b.Factor]-b.Points, m),
			Explanation:  b.Explanation,
		})
	}
	return score, pd, factors
}

// ScoreDisplay — балл в шкале 0–100 для колонки «БАЛЛ».
func ScoreDisplay(score int, m domain.Scorecard) int {
	return int(math.Round(100 * float64(score-m.ScoreMin) / float64(m.ScoreMax-m.ScoreMin)))
}

func binMatches(b domain.Bin, values map[string]float64, industry string) bool {
	if b.Factor == categoricalFactor {
		return b.Category != nil && *b.Category == industry
	}
	v, ok := values[b.Factor]
	return ok && b.Matches(v)
}

func bestPoints(bins []domain.Bin) map[string]int {
	best := map[string]int{}
	for _, b := range bins {
		if cur, ok := best[b.Factor]; !ok || b.Points > cur {
			best[b.Factor] = b.Points
		}
	}
	return best
}

func valueLabel(factor string, values map[string]float64, industry string) string {
	if factor == categoricalFactor {
		return industry
	}
	return fmt.Sprintf("%g", values[factor])
}
