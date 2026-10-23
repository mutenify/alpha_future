package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"smartcredit/internal/domain"
	"smartcredit/internal/usecase/workflow"
)

// Sign — ПЭП-симуляция: APPROVED → SIGNED, выдачу делает шаг DISBURSE конвейера.
func (s *Service) Sign(ctx context.Context, p domain.Principal, id, code string) (domain.Contract, domain.Status, error) {
	if code == "" {
		return domain.Contract{}, "", domain.Invalid("введите код подтверждения")
	}
	app, err := s.load(ctx, p, id)
	if err != nil {
		return domain.Contract{}, "", err
	}
	if err := requireStatus(app, domain.StatusApproved); err != nil {
		return domain.Contract{}, "", err
	}
	offer, err := s.signedOffer(ctx, app.ID)
	if err != nil {
		return domain.Contract{}, "", err
	}
	var (
		contract domain.Contract
		change   domain.StatusChange
	)
	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		now := time.Now()
		contract, err = s.Contracts.Create(ctx, domain.Contract{
			ApplicationID: app.ID, OfferID: offer.ID,
			ContractNumber:    fmt.Sprintf("KD-%d", app.Number),
			SignedAt:          &now,
			SignatureHash:     signature(app.ID, offer.ID, code, now),
			SettlementAccount: app.Company.SettlementAccount,
		})
		if err != nil {
			return err
		}
		if change, err = s.Flow.Move(ctx, &app, domain.StatusSigned, workflow.By(p), ""); err != nil {
			return err
		}
		return s.Jobs.Enqueue(ctx, app.ID, domain.StepDisburse)
	})
	if err != nil {
		return domain.Contract{}, "", err
	}
	s.Flow.Publish(change)
	return contract, app.Status, nil
}

// signedOffer — принятый вариант или исходное предложение при автоодобрении.
func (s *Service) signedOffer(ctx context.Context, appID string) (domain.Offer, error) {
	offers, err := s.Offers.ListByApplication(ctx, appID)
	if err != nil {
		return domain.Offer{}, err
	}
	for _, o := range offers {
		if o.Status == domain.OfferAccepted || (o.Kind == domain.OfferOriginal && o.Status == domain.OfferProposed) {
			return o, nil
		}
	}
	return domain.Offer{}, domain.ErrConflict
}

func signature(appID, offerID, code string, at time.Time) string {
	h := sha256.Sum256([]byte(appID + offerID + code + at.Format(time.RFC3339Nano)))
	return hex.EncodeToString(h[:])
}
