package pipeline

import (
	"context"

	"smartcredit/internal/domain"
	"smartcredit/internal/scoring"
)

// score — слой 1 (PD) и слой 2 (кассовый разрыв), затем решение SYSTEM:
// APPROVED, COUNTER_OFFER (точки А и Б), MANUAL_REVIEW или REJECTED.
func (s *Service) score(ctx context.Context, app domain.Application) error {
	card, err := s.Rules.ActiveScorecard(ctx)
	if err != nil {
		return err
	}
	profile, err := s.Scores.LatestProfile(ctx, app.ID)
	if err != nil {
		return err
	}
	pr := app.Product
	points, pd, factors := scoring.Score(profile, card)
	rate := scoring.EquilibriumRate(pr.BaseRate, pd, pr.LGD)
	outcome := scoring.Route(pd, pr, scoring.Fits(profile, pr, app.Amount, app.TermMonths, rate))
	if outcome == domain.OutcomeCounterOffer && len(variants(app, outcome, profile, rate)) == 0 {
		outcome = domain.OutcomeManualReview // ни А, ни Б не проходят — решает инспектор
	}
	text, source := s.Explainer.Explain(ctx, outcome, factors)

	return s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		res, err := s.Scores.SaveResult(ctx, domain.ScoreResult{
			ApplicationID: app.ID, ProfileID: profile.ID, ModelVersion: card.Version,
			Score: points, PD: pd, Factors: factors, EquilibriumRate: rate,
		})
		if err != nil {
			return err
		}
		dec, err := s.Decisions.Create(ctx, domain.Decision{
			ApplicationID: app.ID, ScoringResultID: res.ID, DecidedBy: domain.ActorSystem,
			Outcome: outcome, ReasonText: text, ExplanationSource: source,
		})
		if err != nil {
			return err
		}
		if err := s.createOffers(ctx, app, dec, profile, rate); err != nil {
			return err
		}
		change, err := s.Flow.Move(ctx, &app, outcome.Status(), systemActor, text)
		if err == nil {
			defer s.Flow.Publish(change)
		}
		return err
	})
}
