package inspector

import (
	"context"

	"smartcredit/internal/domain"
	"smartcredit/internal/scoring"
)

// Scoring — блоки карточки заявки: PD и факторы, слой 2, данные источников.
func (s *Service) Scoring(ctx context.Context, appID string) (domain.ScoringDetails, error) {
	app, err := s.Apps.Get(ctx, appID)
	if err != nil {
		return domain.ScoringDetails{}, err
	}
	var d domain.ScoringDetails
	if d.Result, err = s.Scores.LatestResult(ctx, appID); err != nil {
		return d, err
	}
	if d.Profile, err = s.Scores.LatestProfile(ctx, appID); err != nil {
		return d, err
	}
	if d.Enrichment, err = s.Scores.Enrichment(ctx, appID); err != nil {
		return d, err
	}
	rate := rateFor(app.Product, &d.Result)
	gap := scoring.CashGapMetrics(d.Profile, app.Product, scoring.Payment(app.Amount, app.TermMonths, rate), app.TermMonths)
	d.CashGap = &gap
	return d, nil
}

func rateFor(p domain.Product, r *domain.ScoreResult) float64 {
	if r != nil && r.EquilibriumRate > 0 {
		return r.EquilibriumRate
	}
	return p.BaseRate
}
