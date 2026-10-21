package domain

import "time"

type AuditEntry struct {
	ID          int64
	ActorType   ActorType
	ActorUserID string
	ActorName   string
	Action      string
	EntityType  string
	EntityID    string
	Message     string
	Payload     map[string]any
	CreatedAt   time.Time
}

type AuditFilter struct {
	ApplicationID string
	ActorType     ActorType
	Limit         int
	Before        int64
}
