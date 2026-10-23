package application

import (
	"context"

	"smartcredit/internal/domain"
	"smartcredit/internal/usecase/workflow"
)

// AcceptOffer — клиент выбрал вариант А или Б: COUNTER_OFFER → APPROVED.
func (s *Service) AcceptOffer(ctx context.Context, p domain.Principal, id, offerID string) (domain.ApplicationView, error) {
	return s.resolveOffer(ctx, p, id, domain.StatusApproved, func(ctx context.Context) error {
		_, err := s.Offers.Accept(ctx, id, offerID)
		return err
	})
}

// DeclineOffer — COUNTER_OFFER → CANCELLED.
func (s *Service) DeclineOffer(ctx context.Context, p domain.Principal, id string) (domain.ApplicationView, error) {
	return s.resolveOffer(ctx, p, id, domain.StatusCancelled, func(ctx context.Context) error {
		return s.Offers.DeclineAll(ctx, id)
	})
}

func (s *Service) resolveOffer(ctx context.Context, p domain.Principal, id string, to domain.Status, apply func(context.Context) error) (domain.ApplicationView, error) {
	app, err := s.load(ctx, p, id)
	if err != nil {
		return domain.ApplicationView{}, err
	}
	if err := requireStatus(app, domain.StatusCounterOffer); err != nil {
		return domain.ApplicationView{}, err
	}
	var change domain.StatusChange
	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := apply(ctx); err != nil {
			return err
		}
		change, err = s.Flow.Move(ctx, &app, to, workflow.By(p), "")
		return err
	})
	if err != nil {
		return domain.ApplicationView{}, err
	}
	s.Flow.Publish(change)
	return s.View(ctx, app)
}
