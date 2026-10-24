package handler

import (
	"net/http"

	"smartcredit/internal/transport/http/httpx"
)

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	httpx.OK(w, map[string]string{"status": "ok"})
}

func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	checks := h.Ready(r.Context())
	status, code := "ok", http.StatusOK
	for _, check := range checks {
		if check != "ok" {
			status, code = "down", http.StatusServiceUnavailable
		}
	}
	httpx.JSON(w, code, map[string]any{"status": status, "checks": checks})
}
