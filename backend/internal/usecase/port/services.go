package port

import (
	"context"
	"io"

	"smartcredit/internal/domain"
)

type FileStorage interface {
	Save(ctx context.Context, key string, r io.Reader) (int64, error)
}

type ExplainInput struct {
	Outcome domain.Outcome
	Score   int
	PD      float64
	PGap    float64
	Factors []domain.Factor
}

// Explainer пишет объяснение решения простым языком (LLM с шаблонным фолбэком).
type Explainer interface {
	Explain(ctx context.Context, input ExplainInput) (text, source string)
}

type ProductRecommender interface {
	Recommend(ctx context.Context, purpose string, amount float64, termMonths int) (code, explanation, source string, err error)
}

// EventPublisher рассылает смену статуса подписчикам SSE.
type EventPublisher interface {
	Publish(c domain.StatusChange)
}

type EventSubscriber interface {
	Subscribe(appID string) (ch <-chan domain.StatusChange, cancel func())
}
