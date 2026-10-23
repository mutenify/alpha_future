package application

import (
	"context"

	"smartcredit/internal/domain"
)

func (s *Service) UpdateDraft(ctx context.Context, p domain.Principal, id string, in domain.ApplicationInput) (domain.ApplicationView, error) {
	app, err := s.load(ctx, p, id)
	if err != nil {
		return domain.ApplicationView{}, err
	}
	if app.Status != domain.StatusDraft {
		return domain.ApplicationView{}, domain.ErrConflict
	}
	if app.Product, err = s.Products.ByCode(ctx, in.ProductCode); err != nil {
		return domain.ApplicationView{}, err
	}
	if err := app.Product.Validate(in.Amount, in.TermMonths); err != nil {
		return domain.ApplicationView{}, err
	}
	if app.Company.INN != in.CompanyINN {
		if app.Company, err = s.companyFor(ctx, p.UserID, in.CompanyINN); err != nil {
			return domain.ApplicationView{}, err
		}
	}
	app.Amount, app.TermMonths, app.Purpose = in.Amount, in.TermMonths, in.Purpose
	if app, err = s.Apps.UpdateDraft(ctx, app); err != nil {
		return domain.ApplicationView{}, err
	}
	return domain.ApplicationView{Application: app, NextStep: nextStep(app.Status, nil)}, nil
}
