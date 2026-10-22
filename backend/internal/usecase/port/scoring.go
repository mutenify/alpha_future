package port

import (
	"context"

	"smartcredit/internal/domain"
)

type ScoringRepository interface {
	SaveEnrichment(ctx context.Context, appID string, r domain.EnrichmentResult) error
	Enrichment(ctx context.Context, appID string) ([]domain.EnrichmentResult, error)
	SaveProfile(ctx context.Context, p domain.FinancialProfile) (domain.FinancialProfile, error)
	LatestProfile(ctx context.Context, appID string) (domain.FinancialProfile, error)
	SaveResult(ctx context.Context, r domain.ScoreResult) (domain.ScoreResult, error)
	LatestResult(ctx context.Context, appID string) (domain.ScoreResult, error)
}

type RulesRepository interface {
	ActiveScorecard(ctx context.Context) (domain.Scorecard, error)
	Industries(ctx context.Context) ([]domain.Industry, error)
	Industry(ctx context.Context, code string) (domain.Industry, error)
}
