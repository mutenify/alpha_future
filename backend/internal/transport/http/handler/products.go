package handler

import (
	"net/http"

	"smartcredit/internal/transport/http/dto"
	"smartcredit/internal/transport/http/httpx"
)

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	ps, err := h.Catalog.List(r.Context())
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.OK(w, dto.FromProducts(ps))
}

func (h *Handler) Calculate(w http.ResponseWriter, r *http.Request) {
	var req dto.LoanRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, r, err)
		return
	}
	c, err := h.Catalog.Calculate(r.Context(), req.ProductCode, req.Amount, req.TermMonths)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.OK(w, dto.FromCalculation(c))
}

func (h *Handler) Recommend(w http.ResponseWriter, r *http.Request) {
	var req dto.LoanRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, r, err)
		return
	}
	rec, err := h.Catalog.Recommend(r.Context(), req.Purpose, req.Amount, req.TermMonths)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.OK(w, dto.FromRecommendation(rec))
}
