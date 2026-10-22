package port

import (
	"context"
	"time"

	"smartcredit/internal/domain"
)

type JobRepository interface {
	Enqueue(ctx context.Context, appID string, step domain.JobStep) error
	// Claim берёт задачу FOR UPDATE SKIP LOCKED; nil — очередь пуста.
	Claim(ctx context.Context, lease time.Duration) (*domain.Job, error)
	Done(ctx context.Context, id int64) error
	Retry(ctx context.Context, id int64, runAt time.Time, errMsg string) error
	Fail(ctx context.Context, id int64, errMsg string) error
}

type StoredResponse struct {
	RequestHash string
	Status      int
	Body        []byte
}

type IdempotencyRepository interface {
	Get(ctx context.Context, userID, key string) (*StoredResponse, error)
	Save(ctx context.Context, userID, key, method, path string, r StoredResponse) error
}
