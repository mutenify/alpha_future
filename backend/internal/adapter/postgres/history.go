package postgres

import (
	"context"

	"smartcredit/internal/domain"
)

type History struct{ *DB }

const historyCols = `h.application_id, coalesce(h.from_status, ''), h.to_status, h.actor_type,
	coalesce(h.actor_user_id::text, ''), coalesce(h.reason, ''), h.created_at`

func (r History) Add(ctx context.Context, c domain.StatusChange) error {
	_, err := r.q(ctx).Exec(ctx, `
		INSERT INTO application_status_history (application_id, from_status, to_status, actor_type, actor_user_id, reason)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		c.ApplicationID, nullable(string(c.From)), c.To, c.ActorType, nullable(c.ActorUserID), nullable(c.Reason))
	return err
}

func (r History) ListByApplication(ctx context.Context, appID string) ([]domain.StatusChange, error) {
	rows, err := r.q(ctx).Query(ctx, `SELECT `+historyCols+` FROM application_status_history h
		WHERE h.application_id = $1 ORDER BY h.created_at`, appID)
	if err != nil {
		return nil, err
	}
	return collect(rows, scanChange)
}

func (r History) RecentByClient(ctx context.Context, userID string, limit int) ([]domain.StatusChange, error) {
	rows, err := r.q(ctx).Query(ctx, `SELECT `+historyCols+` FROM application_status_history h
		JOIN applications a ON a.id = h.application_id
		WHERE a.client_user_id = $1 ORDER BY h.created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	return collect(rows, scanChange)
}

func scanChange(row scanner) (domain.StatusChange, error) {
	var c domain.StatusChange
	err := row.Scan(&c.ApplicationID, &c.From, &c.To, &c.ActorType, &c.ActorUserID, &c.Reason, &c.CreatedAt)
	return c, err
}
