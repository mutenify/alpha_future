// Package events — in-memory рассылка смен статуса для SSE (один инстанс Core).
package events

import (
	"sync"

	"smartcredit/internal/domain"
)

const buffer = 16

type Broker struct {
	mu   sync.RWMutex
	subs map[string]map[chan domain.StatusChange]struct{}
}

func NewBroker() *Broker {
	return &Broker{subs: map[string]map[chan domain.StatusChange]struct{}{}}
}

func (b *Broker) Subscribe(appID string) (<-chan domain.StatusChange, func()) {
	ch := make(chan domain.StatusChange, buffer)
	b.mu.Lock()
	if b.subs[appID] == nil {
		b.subs[appID] = map[chan domain.StatusChange]struct{}{}
	}
	b.subs[appID][ch] = struct{}{}
	b.mu.Unlock()
	return ch, func() {
		b.mu.Lock()
		delete(b.subs[appID], ch)
		b.mu.Unlock()
	}
}

// Publish не блокирует: медленный подписчик пропускает событие и дочитает статус через GET.
func (b *Broker) Publish(c domain.StatusChange) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs[c.ApplicationID] {
		select {
		case ch <- c:
		default:
		}
	}
}
