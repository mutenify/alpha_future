package pipeline

import (
	"context"

	"smartcredit/internal/domain"
)

// disburse: SIGNED → DISBURSED через мок АБС.
func (s *Service) disburse(ctx context.Context, app domain.Application) error {
	c, err := s.Contracts.ByApplication(ctx, app.ID)
	if err != nil {
		return err
	}
	amount, err := s.contractAmount(ctx, c)
	if err != nil {
		return err
	}
	if _, err := s.Sources.Disburse(ctx, c.ContractNumber, c.SettlementAccount, amount); err != nil {
		return err
	}
	return s.move(ctx, app, domain.StatusDisbursed, "", func(ctx context.Context) error {
		return s.Contracts.MarkDisbursed(ctx, c.ID)
	})
}

func (s *Service) contractAmount(ctx context.Context, c domain.Contract) (float64, error) {
	offers, err := s.Offers.ListByApplication(ctx, c.ApplicationID)
	if err != nil {
		return 0, err
	}
	for _, o := range offers {
		if o.ID == c.OfferID {
			return o.Amount, nil
		}
	}
	return 0, domain.ErrNotFound
}
