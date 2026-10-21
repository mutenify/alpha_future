package port

import (
	"context"

	"smartcredit/internal/domain"
)

type UserRepository interface {
	ByLogin(ctx context.Context, login string) (domain.User, error)
	ByID(ctx context.Context, id string) (domain.User, error)
	DemoUsers(ctx context.Context) ([]domain.User, error)
}
