package application

import (
	"context"

	"smartcredit/internal/domain"
)

func (s *Service) Get(ctx context.Context, p domain.Principal, id string) (domain.ApplicationView, error) {
	app, err := s.load(ctx, p, id)
	if err != nil {
		return domain.ApplicationView{}, err
	}
	return s.View(ctx, app)
}

func (s *Service) List(ctx context.Context, p domain.Principal, status domain.Status) ([]domain.Application, error) {
	return s.Apps.ListByClient(ctx, p.UserID, status)
}

func (s *Service) Timeline(ctx context.Context, p domain.Principal, id string) ([]domain.StatusChange, error) {
	if _, err := s.load(ctx, p, id); err != nil {
		return nil, err
	}
	return s.History.ListByApplication(ctx, id)
}

// View собирает то, что показывает вкладка «Моя заявка».
func (s *Service) View(ctx context.Context, app domain.Application) (domain.ApplicationView, error) {
	v := domain.ApplicationView{Application: app}
	var err error
	if v.Offers, err = s.Offers.ListByApplication(ctx, app.ID); err != nil {
		return v, err
	}
	if v.OpenInfoRequest, err = s.Infos.Open(ctx, app.ID); err != nil {
		return v, err
	}
	d, err := s.Decisions.Latest(ctx, app.ID)
	if err != nil {
		return v, err
	}
	if d != nil {
		v.DecisionReason = d.ReasonText
	}
	v.NextStep = nextStep(app.Status, v.OpenInfoRequest)
	return v, nil
}
