package dto

import (
	"time"

	"smartcredit/internal/domain"
)

type AuditEntry struct {
	ID         int64            `json:"id"`
	ActorType  domain.ActorType `json:"actorType"`
	ActorName  string           `json:"actorName,omitempty"`
	Action     string           `json:"action"`
	EntityType string           `json:"entityType"`
	EntityID   string           `json:"entityId,omitempty"`
	Message    string           `json:"message,omitempty"`
	CreatedAt  time.Time        `json:"createdAt"`
}

func FromAudit(es []domain.AuditEntry) []AuditEntry {
	return mapAll(es, func(e domain.AuditEntry) AuditEntry {
		return AuditEntry{e.ID, e.ActorType, e.ActorName, e.Action, e.EntityType, e.EntityID, e.Message, e.CreatedAt}
	})
}
