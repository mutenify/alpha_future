package postgres

import (
	"context"

	"smartcredit/internal/domain"
)

type Scoring struct{ *DB }

const profileCols = `id, application_id, monthly_revenue, coalesce(revenue_cv, 0), coalesce(revenue_trend, 0),
	cash_balance, current_debt_payments, tax_debt, bki_overdues_12m, max_overdue_days, bki_inquiries_3m,
	business_age_months, coalesce(okved, ''), coalesce(industry_code, ''), coalesce(industry_margin, 0),
	cash_flow_mu, cash_flow_sigma, coalesce(stop_factor, '')`

func (r Scoring) SaveProfile(ctx context.Context, p domain.FinancialProfile) (domain.FinancialProfile, error) {
	return scanProfile(r.q(ctx).QueryRow(ctx, `
		INSERT INTO financial_profiles (application_id, monthly_revenue, revenue_cv, revenue_trend, cash_balance,
			current_debt_payments, tax_debt, bki_overdues_12m, max_overdue_days, bki_inquiries_3m, business_age_months,
			okved, industry_code, industry_margin, cash_flow_mu, cash_flow_sigma, stop_factor)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		RETURNING `+profileCols,
		p.ApplicationID, mapSlice(p.MonthlyRevenue, toMonthJSON), p.RevenueCV, p.RevenueTrend, p.CashBalance,
		p.CurrentDebtPayments, p.TaxDebt, p.Overdues12m, p.MaxOverdueDays, p.Inquiries3m, p.BusinessAgeMonths,
		nullable(p.OKVED), nullable(p.IndustryCode), p.IndustryMargin, p.CashFlowMu, p.CashFlowSigma, nullable(p.StopFactor)))
}

func (r Scoring) LatestProfile(ctx context.Context, appID string) (domain.FinancialProfile, error) {
	p, err := scanProfile(r.q(ctx).QueryRow(ctx, `SELECT `+profileCols+` FROM financial_profiles
		WHERE application_id = $1 ORDER BY created_at DESC LIMIT 1`, appID))
	return p, notFound(err)
}

func scanProfile(row scanner) (domain.FinancialProfile, error) {
	var (
		p   domain.FinancialProfile
		rev []monthJSON
	)
	err := row.Scan(&p.ID, &p.ApplicationID, &rev, &p.RevenueCV, &p.RevenueTrend, &p.CashBalance,
		&p.CurrentDebtPayments, &p.TaxDebt, &p.Overdues12m, &p.MaxOverdueDays, &p.Inquiries3m,
		&p.BusinessAgeMonths, &p.OKVED, &p.IndustryCode, &p.IndustryMargin, &p.CashFlowMu, &p.CashFlowSigma, &p.StopFactor)
	p.MonthlyRevenue = mapSlice(rev, fromMonthJSON)
	return p, err
}

const resultCols = `id, application_id, financial_profile_id, model_version, score, pd, factors,
	coalesce(equilibrium_rate, 0), coalesce(inspector_metrics, '[]'::jsonb)`

func (r Scoring) SaveResult(ctx context.Context, s domain.ScoreResult) (domain.ScoreResult, error) {
	return scanResult(r.q(ctx).QueryRow(ctx, `
		INSERT INTO scoring_results (application_id, financial_profile_id, model_version, score, pd, factors, equilibrium_rate, inspector_metrics)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING `+resultCols,
		s.ApplicationID, s.ProfileID, s.ModelVersion, s.Score, s.PD,
		mapSlice(s.Factors, toFactorJSON), s.EquilibriumRate, mapSlice(s.InspectorMetrics, toMetricJSON)))
}

func (r Scoring) LatestResult(ctx context.Context, appID string) (domain.ScoreResult, error) {
	s, err := scanResult(r.q(ctx).QueryRow(ctx, `SELECT `+resultCols+` FROM scoring_results
		WHERE application_id = $1 ORDER BY created_at DESC LIMIT 1`, appID))
	return s, notFound(err)
}

func scanResult(row scanner) (domain.ScoreResult, error) {
	var (
		s       domain.ScoreResult
		factors []factorJSON
		metrics []metricJSON
	)
	err := row.Scan(&s.ID, &s.ApplicationID, &s.ProfileID, &s.ModelVersion, &s.Score, &s.PD, &factors,
		&s.EquilibriumRate, &metrics)
	s.Factors = mapSlice(factors, fromFactorJSON)
	s.InspectorMetrics = mapSlice(metrics, fromMetricJSON)
	return s, err
}

func (r Scoring) SaveEnrichment(ctx context.Context, appID string, e domain.EnrichmentResult) error {
	_, err := r.q(ctx).Exec(ctx, `
		INSERT INTO enrichment_results (application_id, source, status, payload, received_at)
		VALUES ($1, $2, $3, $4, now())`, appID, e.Source, e.Status, e.Payload)
	return err
}

func (r Scoring) Enrichment(ctx context.Context, appID string) ([]domain.EnrichmentResult, error) {
	rows, err := r.q(ctx).Query(ctx, `
		SELECT DISTINCT ON (source) source, status, coalesce(payload, '{}'::jsonb)
		FROM enrichment_results WHERE application_id = $1 ORDER BY source, requested_at DESC`, appID)
	if err != nil {
		return nil, err
	}
	return collect(rows, func(row scanner) (domain.EnrichmentResult, error) {
		var e domain.EnrichmentResult
		err := row.Scan(&e.Source, &e.Status, &e.Payload)
		return e, err
	})
}
