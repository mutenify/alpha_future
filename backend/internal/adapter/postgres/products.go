package postgres

import (
	"context"

	"smartcredit/internal/domain"
)

type Products struct{ *DB }

const productCols = `p.id, p.code, p.name, coalesce(p.description, ''), p.min_amount, p.max_amount,
	p.min_term_months, p.max_term_months, p.base_rate, p.lgd, p.pd_auto_approve_max, p.pd_reject_min,
	p.gap_threshold_q, p.payment_share_k`

func (r Products) ListActive(ctx context.Context) ([]domain.Product, error) {
	rows, err := r.q(ctx).Query(ctx, `SELECT `+productCols+` FROM credit_products p WHERE p.is_active ORDER BY p.min_term_months`)
	if err != nil {
		return nil, err
	}
	return collect(rows, scanProduct)
}

func (r Products) ByCode(ctx context.Context, code string) (domain.Product, error) {
	p, err := scanProduct(r.q(ctx).QueryRow(ctx, `SELECT `+productCols+` FROM credit_products p WHERE p.code = $1 AND p.is_active`, code))
	if err != nil {
		return p, notFoundAs(err, domain.Invalid("неизвестный продукт"))
	}
	return p, nil
}

func scanProduct(row scanner) (domain.Product, error) {
	var p domain.Product
	err := row.Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.MinAmount, &p.MaxAmount,
		&p.MinTermMonths, &p.MaxTermMonths, &p.BaseRate, &p.LGD, &p.PDAutoApproveMax, &p.PDRejectMin,
		&p.GapThresholdQ, &p.PaymentShareK)
	return p, err
}
