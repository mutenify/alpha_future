package application

import (
	"context"
	"errors"

	"smartcredit/internal/domain"
	"smartcredit/internal/usecase/company"
)

func (s *Service) Create(ctx context.Context, p domain.Principal, in domain.ApplicationInput) (domain.ApplicationView, error) {
	product, err := s.Products.ByCode(ctx, in.ProductCode)
	if err != nil {
		return domain.ApplicationView{}, err
	}
	if err := product.Validate(in.Amount, in.TermMonths); err != nil {
		return domain.ApplicationView{}, err
	}
	comp, err := s.companyFor(ctx, p.UserID, in.CompanyINN)
	if err != nil {
		return domain.ApplicationView{}, err
	}
	var app domain.Application
	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		app, err = s.Apps.Create(ctx, domain.Application{
			ClientUserID: p.UserID, Company: comp, Product: product,
			Amount: in.Amount, TermMonths: in.TermMonths, Purpose: in.Purpose,
			Status: domain.StatusDraft,
		})
		if err != nil {
			return err
		}
		return s.History.Add(ctx, domain.StatusChange{
			ApplicationID: app.ID, To: domain.StatusDraft, ActorType: domain.ActorClient, ActorUserID: p.UserID,
		})
	})
	return domain.ApplicationView{Application: app, NextStep: nextStep(app.Status, nil)}, err
}

// companyFor берёт компанию клиента по ИНН или создаёт её из ЕГРЮЛ.
func (s *Service) companyFor(ctx context.Context, userID, inn string) (domain.Company, error) {
	c, err := s.Companies.ByINN(ctx, inn)
	if err == nil {
		if c.OwnerUserID != userID {
			return domain.Company{}, domain.ErrForbidden
		}
		return c, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return domain.Company{}, err
	}
	e, err := s.Sources.Egrul(ctx, inn)
	if err != nil {
		return domain.Company{}, err
	}
	c = company.FromEgrul(e)
	c.OwnerUserID = userID
	return s.Companies.Upsert(ctx, c)
}
