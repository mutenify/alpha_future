package port

import (
	"context"
	"io"

	"smartcredit/internal/domain"
)

type FileStorage interface {
	Save(ctx context.Context, key string, r io.Reader) (int64, error)
}

// Explainer пишет объяснение решения простым языком (LLM с шаблонным фолбэком).
type Explainer interface {
	Explain(ctx context.Context, outcome domain.Outcome, factors []domain.Factor) (text, source string)
}

// EventPublisher рассылает смену статуса подписчикам SSE.
type EventPublisher interface {
	Publish(c domain.StatusChange)
}

type EventSubscriber interface {
	Subscribe(appID string) (ch <-chan domain.StatusChange, cancel func())
}
