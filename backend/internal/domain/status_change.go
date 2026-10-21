package domain

import "time"

type StatusChange struct {
	ApplicationID string
	From          Status
	To            Status
	ActorType     ActorType
	ActorUserID   string
	Reason        string
	CreatedAt     time.Time
}
