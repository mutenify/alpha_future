package postgres

import (
	"context"

	"smartcredit/internal/domain"
)

type Audit struct{ *DB }

func (r Audit) Add(ctx context.Context, e domain.AuditEntry) error {
	_, err := r.q(ctx).Exec(ctx, `
		INSERT INTO audit_log (actor_user_id, actor_type, action, entity_type, entity_id, message, payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		nullable(e.ActorUserID), e.ActorType, e.Action, e.EntityType, nullable(e.EntityID), nullable(e.Message), e.Payload)
	return err
}

func (r Audit) List(ctx context.Context, f domain.AuditFilter) ([]domain.AuditEntry, error) {
	rows, err := r.q(ctx).Query(ctx, `
		SELECT l.id, l.actor_type, coalesce(l.actor_user_id::text, ''), coalesce(u.full_name, ''), l.action,
			l.entity_type, coalesce(l.entity_id::text, ''), coalesce(l.message, ''), l.created_at
		FROM audit_log l LEFT JOIN users u ON u.id = l.actor_user_id
		WHERE ($1 = '' OR l.entity_id::text = $1)
		  AND ($2 = '' OR l.actor_type = $2)
		  AND ($3 = 0 OR l.id < $3)
		ORDER BY l.id DESC LIMIT $4`, f.ApplicationID, string(f.ActorType), f.Before, f.Limit)
	if err != nil {
		return nil, err
	}
	return collect(rows, func(row scanner) (domain.AuditEntry, error) {
		var e domain.AuditEntry
		err := row.Scan(&e.ID, &e.ActorType, &e.ActorUserID, &e.ActorName, &e.Action, &e.EntityType, &e.EntityID, &e.Message, &e.CreatedAt)
		return e, err
	})
}
