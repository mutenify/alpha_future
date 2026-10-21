package port

import (
	"context"

	"smartcredit/internal/domain"
)

type ProductRepository interface {
	ListActive(ctx context.Context) ([]domain.Product, error)
	ByCode(ctx context.Context, code string) (domain.Product, error)
}
