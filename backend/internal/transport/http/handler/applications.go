package handler

import (
	"net/http"

	"smartcredit/internal/domain"
	"smartcredit/internal/transport/http/dto"
	"smartcredit/internal/transport/http/httpx"
)

func (h *Handler) ListApplications(w http.ResponseWriter, r *http.Request) {
	status := domain.Status(r.URL.Query().Get("status"))
	apps, err := h.Apps.List(r.Context(), httpx.Principal(r.Context()), status)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.OK(w, dto.FromApplications(apps))
}

func (h *Handler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	in, err := decodeInput(r)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	v, err := h.Apps.Create(r.Context(), httpx.Principal(r.Context()), in)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, http.StatusCreated, dto.FromView(v))
}

func (h *Handler) GetApplication(w http.ResponseWriter, r *http.Request) {
	v, err := h.Apps.Get(r.Context(), httpx.Principal(r.Context()), r.PathValue("id"))
	respondView(w, r, http.StatusOK, v, err)
}

func (h *Handler) UpdateDraft(w http.ResponseWriter, r *http.Request) {
	in, err := decodeInput(r)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	v, err := h.Apps.UpdateDraft(r.Context(), httpx.Principal(r.Context()), r.PathValue("id"), in)
	respondView(w, r, http.StatusOK, v, err)
}

func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	v, err := h.Apps.Submit(r.Context(), httpx.Principal(r.Context()), r.PathValue("id"))
	respondView(w, r, http.StatusAccepted, v, err)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	v, err := h.Apps.Cancel(r.Context(), httpx.Principal(r.Context()), r.PathValue("id"))
	respondView(w, r, http.StatusOK, v, err)
}

func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	changes, err := h.Apps.Timeline(r.Context(), httpx.Principal(r.Context()), r.PathValue("id"))
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.OK(w, dto.FromChanges(changes))
}

func decodeInput(r *http.Request) (domain.ApplicationInput, error) {
	var req dto.ApplicationInput
	if err := httpx.Decode(r, &req); err != nil {
		return domain.ApplicationInput{}, err
	}
	return req.ToDomain()
}

func respondView(w http.ResponseWriter, r *http.Request, status int, v domain.ApplicationView, err error) {
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.JSON(w, status, dto.FromView(v))
}
