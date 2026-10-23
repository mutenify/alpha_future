package handler

import (
	"net/http"

	"smartcredit/internal/domain"
	"smartcredit/internal/transport/http/dto"
	"smartcredit/internal/transport/http/httpx"
)

func (h *Handler) Queue(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	items, counters, err := h.Inspector.Queue(r.Context(), domain.QueueFilter{
		Status: domain.Status(q.Get("status")), Query: q.Get("q"),
	})
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.OK(w, dto.FromQueue(items, counters))
}

func (h *Handler) Scoring(w http.ResponseWriter, r *http.Request) {
	d, err := h.Inspector.Scoring(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.OK(w, dto.FromScoring(d))
}

func (h *Handler) Decide(w http.ResponseWriter, r *http.Request) {
	var req dto.DecisionInput
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, r, err)
		return
	}
	app, err := h.Inspector.Decide(r.Context(), httpx.Principal(r.Context()), r.PathValue("id"), req.ToDomain())
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.OK(w, dto.FromApplication(app))
}
