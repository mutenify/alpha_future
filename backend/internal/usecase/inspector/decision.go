package inspector

import (
	"context"

	"smartcredit/internal/domain"
	"smartcredit/internal/scoring"
	"smartcredit/internal/usecase/workflow"
)

// Decide — кнопки «Одобрить / Контрпредложение / Запросить документы / Отказать».
func (s *Service) Decide(ctx context.Context, p domain.Principal, appID string, in domain.InspectorDecision) (domain.Application, error) {
	if err := in.Validate(); err != nil {
		return domain.Application{}, err
	}
	app, err := s.Apps.Get(ctx, appID)
	if err != nil {
		return domain.Application{}, err
	}
	if app.Status != domain.StatusManualReview {
		return domain.Application{}, domain.ErrInvalidTransition
	}
	var change domain.StatusChange
	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		dec, err := s.Decisions.Create(ctx, domain.Decision{
			ApplicationID: app.ID, DecidedBy: domain.ActorInspector, InspectorID: p.UserID,
			Outcome: in.Outcome, ReasonText: in.ReasonText,
		})
		if err != nil {
			return err
		}
		if err := s.Offers.DeclineAll(ctx, app.ID); err != nil { // старые варианты больше не действуют
			return err
		}
		if err := s.applyOutcome(ctx, p, app, dec, in); err != nil {
			return err
		}
		change, err = s.Flow.Move(ctx, &app, in.Outcome.Status(), workflow.By(p), in.ReasonText)
		return err
	})
	if err != nil {
		return domain.Application{}, err
	}
	s.Flow.Publish(change)
	return app, nil
}

func (s *Service) applyOutcome(ctx context.Context, p domain.Principal, app domain.Application, dec domain.Decision, in domain.InspectorDecision) error {
	switch in.Outcome {
	case domain.OutcomeApproved:
		return s.createOffer(ctx, app, dec, app.Product, domain.OfferOriginal, domain.VariantRequested, app.Amount, app.TermMonths)
	case domain.OutcomeCounterOffer:
		pr := app.Product
		if c := in.CounterOffer; c.ProductCode != "" && c.ProductCode != pr.Code {
			var err error
			if pr, err = s.Products.ByCode(ctx, c.ProductCode); err != nil {
				return err
			}
		}
		if err := pr.Validate(in.CounterOffer.Amount, in.CounterOffer.TermMonths); err != nil {
			return err
		}
		return s.createOffer(ctx, app, dec, pr, domain.OfferCounter, domain.VariantInspector, in.CounterOffer.Amount, in.CounterOffer.TermMonths)
	case domain.OutcomeInfoRequested:
		_, err := s.Infos.Create(ctx, domain.InfoRequest{
			ApplicationID: app.ID, Origin: domain.OriginInspector, RequestedBy: p.UserID, Description: in.InfoRequestText,
		})
		return err
	}
	return nil
}

func (s *Service) createOffer(ctx context.Context, app domain.Application, dec domain.Decision, pr domain.Product,
	kind domain.OfferKind, variant domain.OfferVariant, amount float64, term int) error {
	res, err := s.Scores.LatestResult(ctx, app.ID)
	if err != nil {
		return err
	}
	rate := rateFor(pr, &res)
	_, err = s.Offers.Create(ctx, domain.Offer{
		ApplicationID: app.ID, DecisionID: dec.ID, ProductID: pr.ID, ProductCode: pr.Code,
		Kind: kind, Variant: variant, Amount: amount, TermMonths: term, Rate: rate,
		MonthlyPayment: scoring.Round2(scoring.Payment(amount, term, rate)), Status: domain.OfferProposed,
	})
	return err
}
