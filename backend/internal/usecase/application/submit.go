package application

import (
	"context"

	"smartcredit/internal/domain"
	"smartcredit/internal/usecase/workflow"
)

// Submit — DRAFT → SUBMITTED и задача конвейера в pipeline_jobs; дальше асинхронно.
func (s *Service) Submit(ctx context.Context, p domain.Principal, id string) (domain.ApplicationView, error) {
	app, err := s.load(ctx, p, id)
	if err != nil {
		return domain.ApplicationView{}, err
	}
	var change domain.StatusChange
	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		if change, err = s.Flow.Move(ctx, &app, domain.StatusSubmitted, workflow.By(p), ""); err != nil {
			return err
		}
		return s.Jobs.Enqueue(ctx, app.ID, domain.StepCompleteness)
	})
	if err != nil {
		return domain.ApplicationView{}, err
	}
	s.Flow.Publish(change)
	return s.View(ctx, app)
}

func (s *Service) Cancel(ctx context.Context, p domain.Principal, id string) (domain.ApplicationView, error) {
	app, err := s.load(ctx, p, id)
	if err != nil {
		return domain.ApplicationView{}, err
	}
	var change domain.StatusChange
	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		change, err = s.Flow.Move(ctx, &app, domain.StatusCancelled, workflow.By(p), "Клиент отозвал заявку")
		return err
	})
	if err != nil {
		return domain.ApplicationView{}, err
	}
	s.Flow.Publish(change)
	return s.View(ctx, app)
}
