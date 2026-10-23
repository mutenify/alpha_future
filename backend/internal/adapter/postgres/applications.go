package postgres

import (
	"context"

	"smartcredit/internal/domain"
)

type Applications struct{ *DB }

const appSelect = `SELECT a.id, a.number, a.client_user_id, a.amount, a.term_months, coalesce(a.purpose, ''),
	a.status, coalesce(a.assigned_inspector_id::text, ''), a.version, a.created_at, a.submitted_at, a.updated_at,
	` + companyCols + `, ` + productCols + `
	FROM applications a
	JOIN companies c ON c.id = a.company_id
	JOIN credit_products p ON p.id = a.product_id`

func (r Applications) Create(ctx context.Context, a domain.Application) (domain.Application, error) {
	var id string
	err := r.q(ctx).QueryRow(ctx, `
		INSERT INTO applications (client_user_id, company_id, product_id, amount, term_months, purpose)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		a.ClientUserID, a.Company.ID, a.Product.ID, a.Amount, a.TermMonths, nullable(a.Purpose)).Scan(&id)
	if err != nil {
		return a, err
	}
	return r.Get(ctx, id)
}

func (r Applications) UpdateDraft(ctx context.Context, a domain.Application) (domain.Application, error) {
	tag, err := r.q(ctx).Exec(ctx, `
		UPDATE applications SET company_id = $2, product_id = $3, amount = $4, term_months = $5, purpose = $6,
			version = version + 1, updated_at = now()
		WHERE id = $1 AND status = 'DRAFT' AND version = $7`,
		a.ID, a.Company.ID, a.Product.ID, a.Amount, a.TermMonths, nullable(a.Purpose), a.Version)
	if err != nil {
		return a, err
	}
	if tag.RowsAffected() == 0 {
		return a, domain.ErrConflict
	}
	return r.Get(ctx, a.ID)
}

func (r Applications) Get(ctx context.Context, id string) (domain.Application, error) {
	a, err := scanApplication(r.q(ctx).QueryRow(ctx, appSelect+` WHERE a.id = $1`, id))
	return a, notFound(err)
}

func (r Applications) ListByClient(ctx context.Context, userID string, status domain.Status) ([]domain.Application, error) {
	rows, err := r.q(ctx).Query(ctx, appSelect+`
		WHERE a.client_user_id = $1 AND ($2 = '' OR a.status = $2)
		ORDER BY a.created_at DESC`, userID, string(status))
	if err != nil {
		return nil, err
	}
	return collect(rows, scanApplication)
}

func (r Applications) SetStatus(ctx context.Context, id string, to domain.Status, version int) error {
	tag, err := r.q(ctx).Exec(ctx, `
		UPDATE applications SET status = $2, version = version + 1, updated_at = now(),
			submitted_at = CASE WHEN $2 = 'SUBMITTED' THEN coalesce(submitted_at, now()) ELSE submitted_at END
		WHERE id = $1 AND version = $3`, id, string(to), version)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConflict
	}
	return nil
}

func scanApplication(row scanner) (domain.Application, error) {
	var (
		a domain.Application
		c = &a.Company
		p = &a.Product
	)
	err := row.Scan(&a.ID, &a.Number, &a.ClientUserID, &a.Amount, &a.TermMonths, &a.Purpose,
		&a.Status, &a.AssignedInspectorID, &a.Version, &a.CreatedAt, &a.SubmittedAt, &a.UpdatedAt,
		&c.ID, &c.OwnerUserID, &c.INN, &c.OGRN, &c.Name, &c.LegalForm, &c.OKVED, &c.RegisteredAt, &c.SettlementAccount,
		&p.ID, &p.Code, &p.Name, &p.Description, &p.MinAmount, &p.MaxAmount,
		&p.MinTermMonths, &p.MaxTermMonths, &p.BaseRate, &p.LGD, &p.PDAutoApproveMax, &p.PDRejectMin,
		&p.GapThresholdQ, &p.PaymentShareK)
	return a, err
}
