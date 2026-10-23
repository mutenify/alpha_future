package pipeline

import (
	"context"
	"log/slog"
	"math"
	"time"

	"smartcredit/internal/domain"
)

const (
	pollInterval = 500 * time.Millisecond
	lease        = 30 * time.Second
)

type stepFunc func(ctx context.Context, app domain.Application) error

// Run крутит воркер до отмены ctx.
func (s *Service) Run(ctx context.Context) {
	steps := map[domain.JobStep]stepFunc{
		domain.StepCompleteness: s.checkCompleteness,
		domain.StepEnrich:       s.enrich,
		domain.StepScore:        s.score,
		domain.StepDisburse:     s.disburse,
	}
	t := time.NewTicker(pollInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			for s.runOne(ctx, steps) {
			}
		}
	}
}

// runOne выполняет одну задачу; false — очередь пуста или ошибка очереди.
func (s *Service) runOne(ctx context.Context, steps map[domain.JobStep]stepFunc) bool {
	job, err := s.Jobs.Claim(ctx, lease)
	if err != nil || job == nil {
		if err != nil {
			slog.Error("claim job", "err", err)
		}
		return false
	}
	log := slog.With("job", job.ID, "step", job.Step, "app", job.ApplicationID)
	app, err := s.Apps.Get(ctx, job.ApplicationID)
	if err == nil {
		err = steps[job.Step](ctx, app)
	}
	switch {
	case err == nil:
		err = s.Jobs.Done(ctx, job.ID)
	case job.Attempts+1 < job.MaxAttempts:
		log.Warn("step failed, retry", "err", err)
		err = s.Jobs.Retry(ctx, job.ID, time.Now().Add(backoff(job.Attempts)), err.Error())
	default:
		log.Error("step exhausted", "err", err)
		s.onExhausted(ctx, job, app)
		err = s.Jobs.Fail(ctx, job.ID, err.Error())
	}
	if err != nil {
		log.Error("update job", "err", err)
	}
	return true
}

func backoff(attempt int) time.Duration {
	return time.Duration(math.Pow(2, float64(attempt))) * time.Second
}

// onExhausted: источник так и не ответил — заявку смотрит человек (state.mmd).
func (s *Service) onExhausted(ctx context.Context, job *domain.Job, app domain.Application) {
	if job.Step != domain.StepEnrich {
		return
	}
	s.move(ctx, app, domain.StatusManualReview, "Источник данных недоступен")
}

// move — переход без доп. записей, в своей транзакции.
func (s *Service) move(ctx context.Context, app domain.Application, to domain.Status, reason string, then ...func(context.Context) error) error {
	var change domain.StatusChange
	err := s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		if change, err = s.Flow.Move(ctx, &app, to, systemActor, reason); err != nil {
			return err
		}
		for _, f := range then {
			if err := f(ctx); err != nil {
				return err
			}
		}
		return nil
	})
	if err == nil {
		s.Flow.Publish(change)
	}
	return err
}
