package domain

type Role string

const (
	RoleClient    Role = "CLIENT"
	RoleInspector Role = "INSPECTOR"
	RoleAdmin     Role = "ADMIN"
)

type ActorType string

const (
	ActorSystem    ActorType = "SYSTEM"
	ActorClient    ActorType = "CLIENT"
	ActorInspector ActorType = "INSPECTOR"
)

// Principal — кто выполняет действие (из JWT).
type Principal struct {
	UserID string
	Role   Role
}

func (p Principal) Actor() ActorType {
	if p.Role == RoleClient {
		return ActorClient
	}
	return ActorInspector
}
