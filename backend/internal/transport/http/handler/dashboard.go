package handler

import (
	"net/http"

	"smartcredit/internal/transport/http/dto"
	"smartcredit/internal/transport/http/httpx"
)

func (h *Handler) ClientDashboard(w http.ResponseWriter, r *http.Request) {
	d, err := h.Dashboard.Get(r.Context(), httpx.Principal(r.Context()))
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.OK(w, dto.FromDashboard(d))
}
