package audit

import (
	"context"

	"smartcredit/internal/domain"
	"smartcredit/internal/usecase/port"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

type Service struct{ repo port.AuditRepository }

func New(r port.AuditRepository) *Service { return &Service{repo: r} }

// List — вкладка «Журнал действий», новые сверху.
func (s *Service) List(ctx context.Context, f domain.AuditFilter) ([]domain.AuditEntry, error) {
	if f.Limit <= 0 {
		f.Limit = defaultLimit
	}
	f.Limit = min(f.Limit, maxLimit)
	return s.repo.List(ctx, f)
}
