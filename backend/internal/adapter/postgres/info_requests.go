package postgres

import (
	"context"

	"smartcredit/internal/domain"
)

type InfoRequests struct{ *DB }

const infoCols = `id, application_id, origin, coalesce(requested_by::text, ''), description, status`

func (r InfoRequests) Create(ctx context.Context, ir domain.InfoRequest) (domain.InfoRequest, error) {
	return scanInfo(r.q(ctx).QueryRow(ctx, `
		INSERT INTO info_requests (application_id, origin, requested_by, description)
		VALUES ($1, $2, $3, $4) RETURNING `+infoCols,
		ir.ApplicationID, ir.Origin, nullable(ir.RequestedBy), ir.Description))
}

func (r InfoRequests) Open(ctx context.Context, appID string) (*domain.InfoRequest, error) {
	ir, err := scanInfo(r.q(ctx).QueryRow(ctx, `SELECT `+infoCols+` FROM info_requests
		WHERE application_id = $1 AND status = 'OPEN' ORDER BY created_at DESC LIMIT 1`, appID))
	if err = notFound(err); err == domain.ErrNotFound {
		return nil, nil
	}
	return &ir, err
}

func (r InfoRequests) Fulfill(ctx context.Context, id string) error {
	_, err := r.q(ctx).Exec(ctx, `UPDATE info_requests SET status = 'FULFILLED', fulfilled_at = now() WHERE id = $1`, id)
	return err
}

func scanInfo(row scanner) (domain.InfoRequest, error) {
	var ir domain.InfoRequest
	err := row.Scan(&ir.ID, &ir.ApplicationID, &ir.Origin, &ir.RequestedBy, &ir.Description, &ir.Status)
	return ir, err
}
