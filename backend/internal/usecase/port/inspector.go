package port

import (
	"context"

	"smartcredit/internal/domain"
)

type QueueRepository interface {
	List(ctx context.Context, f domain.QueueFilter) ([]domain.QueueItem, error)
	Counters(ctx context.Context) (domain.QueueCounters, error)
}

type AuditRepository interface {
	Add(ctx context.Context, e domain.AuditEntry) error
	List(ctx context.Context, f domain.AuditFilter) ([]domain.AuditEntry, error)
}
