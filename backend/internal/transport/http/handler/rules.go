package handler

import (
	"net/http"

	"smartcredit/internal/transport/http/dto"
	"smartcredit/internal/transport/http/httpx"
)

func (h *Handler) GetRules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.Rules.Get(r.Context())
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.OK(w, dto.FromRules(rules))
}
