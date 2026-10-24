package catalog

import (
	"context"

	"smartcredit/internal/domain"
	"smartcredit/internal/usecase/port"
)

type Service struct {
	products    port.ProductRepository
	recommender port.ProductRecommender
}

func New(p port.ProductRepository, r port.ProductRecommender) *Service {
	return &Service{products: p, recommender: r}
}

func (s *Service) List(ctx context.Context) ([]domain.Product, error) {
	return s.products.ListActive(ctx)
}
