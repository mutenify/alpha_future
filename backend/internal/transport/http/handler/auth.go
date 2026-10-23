package handler

import (
	"net/http"

	"smartcredit/internal/transport/http/dto"
	"smartcredit/internal/transport/http/httpx"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := httpx.Decode(r, &req); err != nil {
		httpx.Error(w, r, err)
		return
	}
	s, err := h.Auth.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.OK(w, dto.LoginResponse{AccessToken: s.Token, Role: s.User.Role, User: dto.FromUser(s.User)})
}

func (h *Handler) DemoUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.Auth.DemoUsers(r.Context())
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.OK(w, dto.FromUsers(users))
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	u, err := h.Auth.Me(r.Context(), httpx.Principal(r.Context()))
	if err != nil {
		httpx.Error(w, r, err)
		return
	}
	httpx.OK(w, dto.FromUser(u))
}
