// Package dashboard — вкладка «Главная» клиента.
package dashboard

import (
	"context"

	"smartcredit/internal/domain"
	"smartcredit/internal/usecase/application"
	"smartcredit/internal/usecase/port"
)

const recentStatusChanges = 5

type Service struct {
	companies port.CompanyRepository
	contracts port.ContractRepository
	history   port.HistoryRepository
	sources   port.Sources
	apps      *application.Service
}

func New(c port.CompanyRepository, k port.ContractRepository, h port.HistoryRepository, s port.Sources, a *application.Service) *Service {
	return &Service{companies: c, contracts: k, history: h, sources: s, apps: a}
}

func (s *Service) Get(ctx context.Context, p domain.Principal) (domain.Dashboard, error) {
	var (
		d   domain.Dashboard
		err error
	)
	if d.Company, err = s.companies.ByOwner(ctx, p.UserID); err != nil {
		return d, err
	}
	d.Account = s.account(ctx, d.Company)
	if d.ActiveLoans, err = s.contracts.ActiveLoans(ctx, p.UserID); err != nil {
		return d, err
	}
	if d.Applications, err = s.openApplications(ctx, p); err != nil {
		return d, err
	}
	d.Notifications, err = s.notifications(ctx, p, d.Applications)
	return d, err
}

// account — остаток из мока АБС; недоступность мока не ломает главную.
func (s *Service) account(ctx context.Context, c domain.Company) *domain.Account {
	accounts, err := s.sources.Accounts(ctx, c.INN)
	if err != nil || len(accounts) == 0 {
		return nil
	}
	for _, a := range accounts {
		if a.Number == c.SettlementAccount {
			return &domain.Account{Number: a.Number, Balance: a.Balance}
		}
	}
	return &domain.Account{Number: accounts[0].Number, Balance: accounts[0].Balance}
}

func (s *Service) openApplications(ctx context.Context, p domain.Principal) ([]domain.ApplicationView, error) {
	apps, err := s.apps.List(ctx, p, "")
	if err != nil {
		return nil, err
	}
	views := make([]domain.ApplicationView, 0, len(apps))
	for _, a := range apps {
		if a.Status.IsTerminal() {
			continue
		}
		v, err := s.apps.View(ctx, a)
		if err != nil {
			return nil, err
		}
		views = append(views, v)
	}
	return views, nil
}
