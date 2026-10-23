package handler

import (
	"net/http"

	"smartcredit/internal/transport/http/dto"
	"smartcredit/internal/transport/http/httpx"
)

// LookupCompany — GET /companies/lookup?inn=
func (h *Handler) LookupCompany(w http.ResponseWriter, r *http.Request) {
	c, err := h.Company.Lookup(r.Context(), r.URL.Query().Get("inn"))
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.OK(w, dto.FromCompany(c))
}
