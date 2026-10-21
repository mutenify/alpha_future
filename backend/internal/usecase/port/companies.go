package port

import (
	"context"

	"smartcredit/internal/domain"
)

type CompanyRepository interface {
	ByINN(ctx context.Context, inn string) (domain.Company, error)
	ByOwner(ctx context.Context, userID string) (domain.Company, error)
	Upsert(ctx context.Context, c domain.Company) (domain.Company, error)
}
