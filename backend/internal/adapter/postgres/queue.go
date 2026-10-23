package postgres

import (
	"context"

	"smartcredit/internal/domain"
)

type Queue struct{ *DB }

func (r Queue) List(ctx context.Context, f domain.QueueFilter) ([]domain.QueueItem, error) {
	rows, err := r.q(ctx).Query(ctx, `
		SELECT id, display_number, status, company_name, inn, product_code, product_name, amount, term_months,
			score, score_display, pd, coalesce(last_decided_by, ''), coalesce(last_outcome, ''), created_at
		FROM v_inspector_queue
		WHERE ($1 = '' OR status = $1)
		  AND ($2 = '' OR company_name ILIKE '%' || $2 || '%' OR inn LIKE $2 || '%' OR display_number = $2)
		ORDER BY (status = 'MANUAL_REVIEW') DESC, created_at DESC`, string(f.Status), f.Query)
	if err != nil {
		return nil, err
	}
	return collect(rows, func(row scanner) (domain.QueueItem, error) {
		var q domain.QueueItem
		err := row.Scan(&q.ID, &q.DisplayNumber, &q.Status, &q.CompanyName, &q.INN, &q.ProductCode, &q.ProductName,
			&q.Amount, &q.TermMonths, &q.Score, &q.ScoreDisplay, &q.PD, &q.LastDecidedBy, &q.LastOutcome, &q.CreatedAt)
		return q, err
	})
}

func (r Queue) Counters(ctx context.Context) (domain.QueueCounters, error) {
	var c domain.QueueCounters
	err := r.q(ctx).QueryRow(ctx, `
		SELECT count(*) FILTER (WHERE status = 'MANUAL_REVIEW'),
		       count(*) FILTER (WHERE status = 'INFO_REQUESTED'),
		       count(*) FILTER (WHERE last_decided_by = 'SYSTEM' AND updated_at >= current_date),
		       count(*)
		FROM v_inspector_queue`).Scan(&c.ManualReview, &c.InfoRequested, &c.AutoDecidedToday, &c.Total)
	return c, err
}
