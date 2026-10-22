package workflow

import (
	"fmt"

	"smartcredit/internal/domain"
)

func statusAudit(app domain.Application, c domain.StatusChange) domain.AuditEntry {
	return domain.AuditEntry{
		ActorType:   c.ActorType,
		ActorUserID: c.ActorUserID,
		Action:      "STATUS_CHANGED",
		EntityType:  "application",
		EntityID:    app.ID,
		Message:     fmt.Sprintf("Заявка Z-%d: %s → %s", app.Number, c.From, c.To),
		Payload:     map[string]any{"from": c.From, "to": c.To, "reason": c.Reason},
	}
}
