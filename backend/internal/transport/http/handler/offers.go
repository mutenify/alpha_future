package handler

import (
	"net/http"

	"smartcredit/internal/transport/http/dto"
	"smartcredit/internal/transport/http/httpx"
)

func (h *Handler) AcceptOffer(w http.ResponseWriter, r *http.Request) {
	var req dto.AcceptOfferRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, r, err)
		return
	}
	v, err := h.Apps.AcceptOffer(r.Context(), httpx.Principal(r.Context()), r.PathValue("id"), req.OfferID)
	respondView(w, r, http.StatusOK, v, err)
}

func (h *Handler) DeclineOffer(w http.ResponseWriter, r *http.Request) {
	v, err := h.Apps.DeclineOffer(r.Context(), httpx.Principal(r.Context()), r.PathValue("id"))
	respondView(w, r, http.StatusOK, v, err)
}

func (h *Handler) Sign(w http.ResponseWriter, r *http.Request) {
	var req dto.SignRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, r, err)
		return
	}
	c, status, err := h.Apps.Sign(r.Context(), httpx.Principal(r.Context()), r.PathValue("id"), req.ConfirmationCode)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.OK(w, dto.FromContract(c, status))
}
