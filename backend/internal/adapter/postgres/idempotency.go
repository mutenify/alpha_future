package postgres

import (
	"context"

	"smartcredit/internal/domain"
	"smartcredit/internal/usecase/port"
)

type Idempotency struct{ *DB }

func (r Idempotency) Get(ctx context.Context, userID, key string) (*port.StoredResponse, error) {
	var s port.StoredResponse
	err := r.q(ctx).QueryRow(ctx, `SELECT request_hash, coalesce(response_status, 0), coalesce(response_body::text, '')
		FROM idempotency_keys WHERE user_id = $1 AND key = $2`, userID, key).Scan(&s.RequestHash, &s.Status, &s.Body)
	if err = notFound(err); err == domain.ErrNotFound {
		return nil, nil
	}
	return &s, err
}

func (r Idempotency) Save(ctx context.Context, userID, key, method, path string, s port.StoredResponse) error {
	body := string(s.Body)
	if body == "" {
		body = "null"
	}
	_, err := r.q(ctx).Exec(ctx, `
		INSERT INTO idempotency_keys (key, user_id, request_method, request_path, request_hash, response_status, response_body)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb) ON CONFLICT DO NOTHING`,
		key, userID, method, path, s.RequestHash, s.Status, body)
	return err
}
