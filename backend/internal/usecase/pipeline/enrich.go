package pipeline

import (
	"context"

	"smartcredit/internal/domain"
	"smartcredit/internal/usecase/port"
)

type sourceData struct {
	egrul    port.EgrulData
	fns      port.FnsData
	bki      port.BkiData
	accounts []port.AbsAccount
}

// enrich: ENRICHING → SCORING или REJECTED (стоп-фактор). Ошибка источника — ретрай задачи.
func (s *Service) enrich(ctx context.Context, app domain.Application) error {
	data, err := s.fetch(ctx, app.Company.INN)
	if err != nil {
		return err
	}
	industry, err := s.Rules.Industry(ctx, app.Company.IndustryCode())
	if err != nil {
		return err
	}
	profile := buildProfile(app, data, industry)
	return s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.saveEnrichment(ctx, app.ID, data); err != nil {
			return err
		}
		if _, err := s.Scores.SaveProfile(ctx, profile); err != nil {
			return err
		}
		if profile.StopFactor != "" {
			return s.rejectByStopFactor(ctx, app, profile.StopFactor)
		}
		change, err := s.Flow.Move(ctx, &app, domain.StatusScoring, systemActor, "")
		if err != nil {
			return err
		}
		defer s.Flow.Publish(change)
		return s.Jobs.Enqueue(ctx, app.ID, domain.StepScore)
	})
}

func (s *Service) fetch(ctx context.Context, inn string) (d sourceData, err error) {
	if d.egrul, err = s.Sources.Egrul(ctx, inn); err != nil {
		return
	}
	if d.fns, err = s.Sources.Fns(ctx, inn); err != nil {
		return
	}
	if d.bki, err = s.Sources.Bki(ctx, inn); err != nil {
		return
	}
	d.accounts, err = s.Sources.Accounts(ctx, inn)
	return
}

func (s *Service) saveEnrichment(ctx context.Context, appID string, d sourceData) error {
	for src, payload := range map[string]any{"EGRUL": d.egrul, "FNS": d.fns, "BKI": d.bki} {
		r := domain.EnrichmentResult{Source: src, Status: "OK", Payload: map[string]any{"data": payload}}
		if err := s.Scores.SaveEnrichment(ctx, appID, r); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) rejectByStopFactor(ctx context.Context, app domain.Application, factor string) error {
	text, source := s.Explainer.Explain(ctx, port.ExplainInput{
		Outcome: domain.OutcomeRejected, Factors: []domain.Factor{{Code: factor}},
	})
	if _, err := s.Decisions.Create(ctx, domain.Decision{
		ApplicationID: app.ID, DecidedBy: domain.ActorSystem, Outcome: domain.OutcomeRejected,
		ReasonCode: factor, ReasonText: text, ExplanationSource: source,
	}); err != nil {
		return err
	}
	change, err := s.Flow.Move(ctx, &app, domain.StatusRejected, systemActor, factor)
	if err == nil {
		defer s.Flow.Publish(change)
	}
	return err
}
