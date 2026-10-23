// Package explain — объяснение решения по шаблонам (фолбэк, когда LLM недоступна).
// TODO: LLM-реализация port.Explainer, которая при ошибке зовёт Template.
package explain

import (
	"context"
	"sort"
	"strings"

	"smartcredit/internal/domain"
)

const topFactors = 2

var headlines = map[domain.Outcome]string{
	domain.OutcomeApproved:     "Заявка одобрена.",
	domain.OutcomeCounterOffer: "Запрошенная сумма выше безопасной для вашего денежного потока, мы подобрали другие условия.",
	domain.OutcomeManualReview: "Заявку дополнительно посмотрит кредитный инспектор.",
	domain.OutcomeRejected:     "К сожалению, сейчас мы не можем выдать кредит.",
}

var stopFactors = map[string]string{
	"LIQUIDATION":      "Компания в процессе ликвидации.",
	"BANKRUPTCY":       "В отношении компании ведётся процедура банкротства.",
	"ACCOUNTS_BLOCKED": "Счета компании заблокированы ФНС.",
}

type Template struct{}

func (Template) Explain(_ context.Context, o domain.Outcome, factors []domain.Factor) (string, string) {
	parts := []string{headlines[o]}
	for _, f := range factors {
		if s, ok := stopFactors[f.Code]; ok {
			parts = append(parts, s)
		}
	}
	parts = append(parts, worstFactors(factors)...)
	return strings.Join(parts, " "), "TEMPLATE"
}

// worstFactors — фразы математика у факторов, сильнее всего поднявших PD.
func worstFactors(factors []domain.Factor) []string {
	sorted := append([]domain.Factor(nil), factors...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Contribution > sorted[j].Contribution })
	var out []string
	for _, f := range sorted {
		if len(out) == topFactors {
			break
		}
		if f.Explanation != "" && f.Contribution > 0 {
			out = append(out, f.Explanation)
		}
	}
	return out
}
