package handler

import (
	"net/http"

	"smartcredit/internal/domain"
	"smartcredit/internal/transport/http/dto"
	"smartcredit/internal/transport/http/httpx"
)

func (h *Handler) AuditLog(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	entries, err := h.Audit.List(r.Context(), domain.AuditFilter{
		ApplicationID: q.Get("applicationId"),
		ActorType:     domain.ActorType(q.Get("actorType")),
		Limit:         int(httpx.QueryInt(r, "limit")),
		Before:        httpx.QueryInt(r, "before"),
	})
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.OK(w, dto.FromAudit(entries))
}
