package port

import (
	"context"

	"smartcredit/internal/domain"
)

type ApplicationRepository interface {
	Create(ctx context.Context, a domain.Application) (domain.Application, error)
	UpdateDraft(ctx context.Context, a domain.Application) (domain.Application, error)
	Get(ctx context.Context, id string) (domain.Application, error)
	ListByClient(ctx context.Context, userID string, status domain.Status) ([]domain.Application, error)
	// SetStatus меняет статус, если version совпала; иначе domain.ErrConflict.
	SetStatus(ctx context.Context, id string, to domain.Status, version int) error
}

type HistoryRepository interface {
	Add(ctx context.Context, c domain.StatusChange) error
	ListByApplication(ctx context.Context, appID string) ([]domain.StatusChange, error)
	RecentByClient(ctx context.Context, userID string, limit int) ([]domain.StatusChange, error)
}

type InfoRequestRepository interface {
	Create(ctx context.Context, r domain.InfoRequest) (domain.InfoRequest, error)
	Open(ctx context.Context, appID string) (*domain.InfoRequest, error)
	Fulfill(ctx context.Context, id string) error
}

type DocumentRepository interface {
	Create(ctx context.Context, d domain.Document) (domain.Document, error)
	ListByApplication(ctx context.Context, appID string) ([]domain.Document, error)
}
