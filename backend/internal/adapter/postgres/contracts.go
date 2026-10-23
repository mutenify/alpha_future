package postgres

import (
	"context"

	"smartcredit/internal/domain"
)

type Contracts struct{ *DB }

const contractCols = `id, application_id, offer_id, contract_number, signed_at, coalesce(signature_hash, ''),
	disbursed_at, coalesce(settlement_account, '')`

func (r Contracts) Create(ctx context.Context, c domain.Contract) (domain.Contract, error) {
	return scanContract(r.q(ctx).QueryRow(ctx, `
		INSERT INTO contracts (application_id, offer_id, contract_number, signed_at, signature_hash, settlement_account)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING `+contractCols,
		c.ApplicationID, c.OfferID, c.ContractNumber, c.SignedAt, c.SignatureHash, nullable(c.SettlementAccount)))
}

func (r Contracts) ByApplication(ctx context.Context, appID string) (domain.Contract, error) {
	c, err := scanContract(r.q(ctx).QueryRow(ctx, `SELECT `+contractCols+` FROM contracts WHERE application_id = $1`, appID))
	return c, notFound(err)
}

func (r Contracts) MarkDisbursed(ctx context.Context, id string) error {
	_, err := r.q(ctx).Exec(ctx, `UPDATE contracts SET disbursed_at = now() WHERE id = $1`, id)
	return err
}

func (r Contracts) ActiveLoans(ctx context.Context, userID string) ([]domain.Loan, error) {
	rows, err := r.q(ctx).Query(ctx, `
		SELECT k.contract_number, p.code, o.amount, o.term_months, o.rate, o.monthly_payment, k.disbursed_at
		FROM contracts k
		JOIN applications a ON a.id = k.application_id
		JOIN offers o ON o.id = k.offer_id
		JOIN credit_products p ON p.id = o.product_id
		WHERE a.client_user_id = $1 AND k.disbursed_at IS NOT NULL
		ORDER BY k.disbursed_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	return collect(rows, func(row scanner) (domain.Loan, error) {
		var l domain.Loan
		err := row.Scan(&l.ContractNumber, &l.ProductCode, &l.Amount, &l.TermMonths, &l.Rate, &l.MonthlyPayment, &l.DisbursedAt)
		return l, err
	})
}

func scanContract(row scanner) (domain.Contract, error) {
	var c domain.Contract
	err := row.Scan(&c.ID, &c.ApplicationID, &c.OfferID, &c.ContractNumber, &c.SignedAt, &c.SignatureHash,
		&c.DisbursedAt, &c.SettlementAccount)
	return c, err
}
