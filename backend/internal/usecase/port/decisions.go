package port

import (
	"context"

	"smartcredit/internal/domain"
)

type DecisionRepository interface {
	Create(ctx context.Context, d domain.Decision) (domain.Decision, error)
	Latest(ctx context.Context, appID string) (*domain.Decision, error)
}

type OfferRepository interface {
	Create(ctx context.Context, o domain.Offer) (domain.Offer, error)
	ListByApplication(ctx context.Context, appID string) ([]domain.Offer, error)
	// Accept помечает оффер ACCEPTED, остальные PROPOSED по заявке — DECLINED.
	Accept(ctx context.Context, appID, offerID string) (domain.Offer, error)
	DeclineAll(ctx context.Context, appID string) error
}

type ContractRepository interface {
	Create(ctx context.Context, c domain.Contract) (domain.Contract, error)
	ByApplication(ctx context.Context, appID string) (domain.Contract, error)
	MarkDisbursed(ctx context.Context, id string) error
	ActiveLoans(ctx context.Context, userID string) ([]domain.Loan, error)
}
