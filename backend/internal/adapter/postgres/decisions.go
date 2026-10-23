package postgres

import (
	"context"

	"smartcredit/internal/domain"
)

type Decisions struct{ *DB }

const decisionCols = `id, application_id, coalesce(scoring_result_id::text, ''), decided_by,
	coalesce(inspector_id::text, ''), outcome, coalesce(reason_code, ''), coalesce(reason_text, ''),
	coalesce(explanation_source, '')`

func (r Decisions) Create(ctx context.Context, d domain.Decision) (domain.Decision, error) {
	return scanDecision(r.q(ctx).QueryRow(ctx, `
		INSERT INTO decisions (application_id, scoring_result_id, decided_by, inspector_id, outcome, reason_code, reason_text, explanation_source)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING `+decisionCols,
		d.ApplicationID, nullable(d.ScoringResultID), d.DecidedBy, nullable(d.InspectorID), d.Outcome,
		nullable(d.ReasonCode), nullable(d.ReasonText), nullable(d.ExplanationSource)))
}

func (r Decisions) Latest(ctx context.Context, appID string) (*domain.Decision, error) {
	d, err := scanDecision(r.q(ctx).QueryRow(ctx, `SELECT `+decisionCols+` FROM decisions
		WHERE application_id = $1 ORDER BY created_at DESC LIMIT 1`, appID))
	if err = notFound(err); err == domain.ErrNotFound {
		return nil, nil
	}
	return &d, err
}

func scanDecision(row scanner) (domain.Decision, error) {
	var d domain.Decision
	err := row.Scan(&d.ID, &d.ApplicationID, &d.ScoringResultID, &d.DecidedBy, &d.InspectorID, &d.Outcome,
		&d.ReasonCode, &d.ReasonText, &d.ExplanationSource)
	return d, err
}
