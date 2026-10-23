package pipeline

import (
	"context"

	"smartcredit/internal/domain"
	"smartcredit/internal/scoring"
)

type variant struct {
	kind  domain.OfferKind
	name  domain.OfferVariant
	terms scoring.Terms
}

// createOffers: для APPROVED — запрошенные условия, для COUNTER_OFFER — точки А и Б.
func (s *Service) createOffers(ctx context.Context, app domain.Application, dec domain.Decision, p domain.FinancialProfile, rate float64) error {
	pr := app.Product
	for _, v := range variants(app, dec.Outcome, p, rate) {
		gap := v.terms.PGap
		if _, err := s.Offers.Create(ctx, domain.Offer{
			ApplicationID: app.ID, DecisionID: dec.ID, ProductID: pr.ID, ProductCode: pr.Code,
			Kind: v.kind, Variant: v.name, Amount: v.terms.Amount, TermMonths: v.terms.TermMonths,
			Rate: rate, MonthlyPayment: v.terms.MonthlyPayment, PGap: &gap, Status: domain.OfferProposed,
		}); err != nil {
			return err
		}
	}
	return nil
}

func variants(app domain.Application, outcome domain.Outcome, p domain.FinancialProfile, rate float64) []variant {
	pr := app.Product
	switch outcome {
	case domain.OutcomeApproved:
		a := scoring.Payment(app.Amount, app.TermMonths, rate)
		return []variant{{domain.OfferOriginal, domain.VariantRequested, scoring.Terms{
			Amount: app.Amount, TermMonths: app.TermMonths, MonthlyPayment: scoring.Round2(a),
			PGap: scoring.PGap(p.CashFlowMu, p.CashFlowSigma, p.CurrentDebtPayments, a),
		}}}
	case domain.OutcomeCounterOffer:
		var out []variant
		if t, ok := scoring.PointA(p, pr, app.TermMonths, rate); ok {
			out = append(out, variant{domain.OfferCounter, domain.VariantPointA, t})
		}
		if t, ok := scoring.PointB(p, pr, app.Amount, app.TermMonths, rate); ok {
			out = append(out, variant{domain.OfferCounter, domain.VariantPointB, t})
		}
		return out
	}
	return nil
}
