package postgres

import (
	"context"

	"smartcredit/internal/domain"
)

type Offers struct{ *DB }

const offerSelect = `SELECT o.id, o.application_id, o.decision_id, o.product_id, p.code, o.kind, o.variant,
	o.amount, o.term_months, o.rate, o.monthly_payment, o.p_gap, o.status, o.expires_at
	FROM offers o JOIN credit_products p ON p.id = o.product_id`

func (r Offers) Create(ctx context.Context, o domain.Offer) (domain.Offer, error) {
	var id string
	err := r.q(ctx).QueryRow(ctx, `
		INSERT INTO offers (application_id, decision_id, product_id, kind, variant, amount, term_months, rate, monthly_payment, p_gap, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`,
		o.ApplicationID, o.DecisionID, o.ProductID, o.Kind, o.Variant, o.Amount, o.TermMonths, o.Rate,
		o.MonthlyPayment, o.PGap, o.ExpiresAt).Scan(&id)
	if err != nil {
		return o, err
	}
	return scanOffer(r.q(ctx).QueryRow(ctx, offerSelect+` WHERE o.id = $1`, id))
}

func (r Offers) ListByApplication(ctx context.Context, appID string) ([]domain.Offer, error) {
	rows, err := r.q(ctx).Query(ctx, offerSelect+` WHERE o.application_id = $1 ORDER BY o.created_at, o.variant`, appID)
	if err != nil {
		return nil, err
	}
	return collect(rows, scanOffer)
}

func (r Offers) Accept(ctx context.Context, appID, offerID string) (domain.Offer, error) {
	tag, err := r.q(ctx).Exec(ctx, `UPDATE offers SET status = 'ACCEPTED'
		WHERE id = $1 AND application_id = $2 AND status = 'PROPOSED'`, offerID, appID)
	if err != nil {
		return domain.Offer{}, err
	}
	if tag.RowsAffected() == 0 {
		return domain.Offer{}, domain.ErrNotFound
	}
	if _, err := r.q(ctx).Exec(ctx, `UPDATE offers SET status = 'DECLINED'
		WHERE application_id = $1 AND status = 'PROPOSED'`, appID); err != nil {
		return domain.Offer{}, err
	}
	return scanOffer(r.q(ctx).QueryRow(ctx, offerSelect+` WHERE o.id = $1`, offerID))
}

func (r Offers) DeclineAll(ctx context.Context, appID string) error {
	_, err := r.q(ctx).Exec(ctx, `UPDATE offers SET status = 'DECLINED' WHERE application_id = $1 AND status = 'PROPOSED'`, appID)
	return err
}

func scanOffer(row scanner) (domain.Offer, error) {
	var o domain.Offer
	err := row.Scan(&o.ID, &o.ApplicationID, &o.DecisionID, &o.ProductID, &o.ProductCode, &o.Kind, &o.Variant,
		&o.Amount, &o.TermMonths, &o.Rate, &o.MonthlyPayment, &o.PGap, &o.Status, &o.ExpiresAt)
	return o, err
}
