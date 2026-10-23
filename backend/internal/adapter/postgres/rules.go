package postgres

import (
	"context"

	"smartcredit/internal/domain"
)

type Rules struct{ *DB }

func (r Rules) ActiveScorecard(ctx context.Context) (domain.Scorecard, error) {
	var m domain.Scorecard
	err := r.q(ctx).QueryRow(ctx, `SELECT version, base_score, base_odds, pdo, score_min, score_max
		FROM scorecard_models WHERE is_active`).
		Scan(&m.Version, &m.BaseScore, &m.BaseOdds, &m.PDO, &m.ScoreMin, &m.ScoreMax)
	if err != nil {
		return m, notFound(err)
	}
	rows, err := r.q(ctx).Query(ctx, `SELECT factor, lower_bound, upper_bound, category, points, coalesce(explanation, '')
		FROM scorecard_bins WHERE model_version = $1 ORDER BY factor, lower_bound NULLS FIRST, category`, m.Version)
	if err != nil {
		return m, err
	}
	m.Bins, err = collect(rows, func(row scanner) (domain.Bin, error) {
		var b domain.Bin
		err := row.Scan(&b.Factor, &b.Lower, &b.Upper, &b.Category, &b.Points, &b.Explanation)
		return b, err
	})
	return m, err
}

func (r Rules) Industries(ctx context.Context) ([]domain.Industry, error) {
	rows, err := r.q(ctx).Query(ctx, `SELECT code, name, margin, revenue_cv FROM industries ORDER BY code`)
	if err != nil {
		return nil, err
	}
	return collect(rows, scanIndustry)
}

// Industry — отрасль по группе ОКВЭД, неизвестная группа даёт OTHER.
func (r Rules) Industry(ctx context.Context, code string) (domain.Industry, error) {
	i, err := scanIndustry(r.q(ctx).QueryRow(ctx, `SELECT code, name, margin, revenue_cv FROM industries
		WHERE code IN ($1, 'OTHER') ORDER BY code = 'OTHER' LIMIT 1`, code))
	return i, notFound(err)
}

func scanIndustry(row scanner) (domain.Industry, error) {
	var i domain.Industry
	err := row.Scan(&i.Code, &i.Name, &i.Margin, &i.RevenueCV)
	return i, err
}
