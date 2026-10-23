package inspector

import (
	"context"

	"smartcredit/internal/domain"
)

func (s *Service) Queue(ctx context.Context, f domain.QueueFilter) ([]domain.QueueItem, domain.QueueCounters, error) {
	items, err := s.QueueRepo.List(ctx, f)
	if err != nil {
		return nil, domain.QueueCounters{}, err
	}
	counters, err := s.QueueRepo.Counters(ctx)
	return items, counters, err
}
