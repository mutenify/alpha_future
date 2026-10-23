package application

import (
	"context"

	"smartcredit/internal/domain"
)

// load достаёт заявку и проверяет, что клиент смотрит свою.
func (s *Service) load(ctx context.Context, p domain.Principal, id string) (domain.Application, error) {
	app, err := s.Apps.Get(ctx, id)
	if err != nil {
		return domain.Application{}, err
	}
	if !app.OwnedBy(p) {
		return domain.Application{}, domain.ErrNotFound
	}
	return app, nil
}

func requireStatus(app domain.Application, allowed ...domain.Status) error {
	for _, s := range allowed {
		if app.Status == s {
			return nil
		}
	}
	return domain.ErrInvalidTransition
}
