package catalog

import (
	"context"

	"smartcredit/internal/domain"
	"smartcredit/internal/usecase/port"
)

type Service struct{ products port.ProductRepository }

func New(p port.ProductRepository) *Service { return &Service{products: p} }

func (s *Service) List(ctx context.Context) ([]domain.Product, error) {
	return s.products.ListActive(ctx)
}
