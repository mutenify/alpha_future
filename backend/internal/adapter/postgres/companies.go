package postgres

import (
	"context"

	"smartcredit/internal/domain"
)

type Companies struct{ *DB }

const companyCols = `c.id, c.owner_user_id, c.inn, coalesce(c.ogrn, ''), c.name, c.legal_form,
	coalesce(c.okved, ''), c.registered_at, coalesce(c.settlement_account, '')`

func (r Companies) ByINN(ctx context.Context, inn string) (domain.Company, error) {
	c, err := scanCompany(r.q(ctx).QueryRow(ctx, `SELECT `+companyCols+` FROM companies c WHERE c.inn = $1`, inn))
	return c, notFound(err)
}

func (r Companies) ByOwner(ctx context.Context, userID string) (domain.Company, error) {
	c, err := scanCompany(r.q(ctx).QueryRow(ctx,
		`SELECT `+companyCols+` FROM companies c WHERE c.owner_user_id = $1 ORDER BY c.created_at LIMIT 1`, userID))
	return c, notFound(err)
}

func (r Companies) Upsert(ctx context.Context, c domain.Company) (domain.Company, error) {
	return scanCompany(r.q(ctx).QueryRow(ctx, `
		INSERT INTO companies AS c (owner_user_id, inn, ogrn, name, legal_form, okved, registered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (inn) DO UPDATE SET ogrn = EXCLUDED.ogrn, name = EXCLUDED.name, okved = EXCLUDED.okved
		RETURNING `+companyCols,
		c.OwnerUserID, c.INN, nullable(c.OGRN), c.Name, c.LegalForm, nullable(c.OKVED), c.RegisteredAt))
}

func scanCompany(row scanner) (domain.Company, error) {
	var c domain.Company
	err := row.Scan(&c.ID, &c.OwnerUserID, &c.INN, &c.OGRN, &c.Name, &c.LegalForm, &c.OKVED, &c.RegisteredAt, &c.SettlementAccount)
	return c, err
}
