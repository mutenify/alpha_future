package postgres

import (
	"context"
	"time"

	"smartcredit/internal/domain"
)

type Jobs struct{ *DB }

func (r Jobs) Enqueue(ctx context.Context, appID string, step domain.JobStep) error {
	_, err := r.q(ctx).Exec(ctx, `INSERT INTO pipeline_jobs (application_id, step) VALUES ($1, $2)
		ON CONFLICT DO NOTHING`, appID, step)
	return err
}

// Claim — см. комментарий в миграции 000003: PENDING по run_at или RUNNING с истёкшей арендой.
func (r Jobs) Claim(ctx context.Context, lease time.Duration) (*domain.Job, error) {
	var j domain.Job
	err := r.q(ctx).QueryRow(ctx, `
		UPDATE pipeline_jobs SET status = 'RUNNING', locked_until = now() + $1::interval, updated_at = now()
		WHERE id = (
			SELECT id FROM pipeline_jobs
			WHERE (status = 'PENDING' AND run_at <= now()) OR (status = 'RUNNING' AND locked_until < now())
			ORDER BY run_at FOR UPDATE SKIP LOCKED LIMIT 1)
		RETURNING id, application_id, step, attempts, max_attempts`, lease.String()).
		Scan(&j.ID, &j.ApplicationID, &j.Step, &j.Attempts, &j.MaxAttempts)
	if err = notFound(err); err == domain.ErrNotFound {
		return nil, nil
	}
	return &j, err
}

func (r Jobs) Done(ctx context.Context, id int64) error {
	_, err := r.q(ctx).Exec(ctx, `UPDATE pipeline_jobs SET status = 'DONE', updated_at = now() WHERE id = $1`, id)
	return err
}

func (r Jobs) Retry(ctx context.Context, id int64, runAt time.Time, errMsg string) error {
	_, err := r.q(ctx).Exec(ctx, `UPDATE pipeline_jobs SET status = 'PENDING', attempts = attempts + 1,
		run_at = $2, last_error = $3, updated_at = now() WHERE id = $1`, id, runAt, errMsg)
	return err
}

func (r Jobs) Fail(ctx context.Context, id int64, errMsg string) error {
	_, err := r.q(ctx).Exec(ctx, `UPDATE pipeline_jobs SET status = 'FAILED', attempts = attempts + 1,
		last_error = $2, updated_at = now() WHERE id = $1`, id, errMsg)
	return err
}
