// Package workflow — единственное место, где меняется статус заявки:
// проверка по машине состояний, запись истории и аудита.
package workflow

import (
	"context"
	"fmt"
	"time"

	"smartcredit/internal/domain"
	"smartcredit/internal/usecase/port"
)

type Transitioner struct {
	apps    port.ApplicationRepository
	history port.HistoryRepository
	audit   port.AuditRepository
	events  port.EventPublisher
}

func New(apps port.ApplicationRepository, h port.HistoryRepository, a port.AuditRepository, e port.EventPublisher) *Transitioner {
	return &Transitioner{apps: apps, history: h, audit: a, events: e}
}

// Actor — кто двигает заявку.
type Actor struct {
	Type   domain.ActorType
	UserID string
}

func System() Actor               { return Actor{Type: domain.ActorSystem} }
func By(p domain.Principal) Actor { return Actor{Type: p.Actor(), UserID: p.UserID} }

// Move вызывается внутри транзакции. Событие для SSE отдаётся вызывающему,
// он публикует его через Publish после коммита.
func (t *Transitioner) Move(ctx context.Context, app *domain.Application, to domain.Status, by Actor, reason string) (domain.StatusChange, error) {
	if !domain.CanTransition(app.Status, to) {
		return domain.StatusChange{}, fmt.Errorf("%w: %s → %s", domain.ErrInvalidTransition, app.Status, to)
	}
	if err := t.apps.SetStatus(ctx, app.ID, to, app.Version); err != nil {
		return domain.StatusChange{}, err
	}
	change := domain.StatusChange{
		ApplicationID: app.ID, From: app.Status, To: to,
		ActorType: by.Type, ActorUserID: by.UserID, Reason: reason, CreatedAt: time.Now(),
	}
	if err := t.history.Add(ctx, change); err != nil {
		return domain.StatusChange{}, err
	}
	if err := t.audit.Add(ctx, statusAudit(*app, change)); err != nil {
		return domain.StatusChange{}, err
	}
	app.Status, app.Version = to, app.Version+1
	return change, nil
}

func (t *Transitioner) Publish(changes ...domain.StatusChange) {
	for _, c := range changes {
		if c.To != "" {
			t.events.Publish(c)
		}
	}
}
