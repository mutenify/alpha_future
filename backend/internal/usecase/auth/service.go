package auth

import (
	"context"
	"errors"

	"smartcredit/internal/domain"
	"smartcredit/internal/usecase/port"
)

type Service struct {
	users     port.UserRepository
	tokens    port.TokenIssuer
	passwords port.PasswordChecker
	demoMode  bool
}

func New(u port.UserRepository, t port.TokenIssuer, p port.PasswordChecker, demoMode bool) *Service {
	return &Service{users: u, tokens: t, passwords: p, demoMode: demoMode}
}

type Session struct {
	Token string
	User  domain.User
}

// Login: в демо-режиме у демо-учёток пароль не проверяется (кнопки на экране входа).
func (s *Service) Login(ctx context.Context, login, password string) (Session, error) {
	u, err := s.users.ByLogin(ctx, login)
	if errors.Is(err, domain.ErrNotFound) {
		return Session{}, domain.ErrUnauthorized
	}
	if err != nil {
		return Session{}, err
	}
	if !(s.demoMode && u.ScenarioHint != "") && !s.passwords.Check(u.PasswordHash, password) {
		return Session{}, domain.ErrUnauthorized
	}
	token, err := s.tokens.Issue(domain.Principal{UserID: u.ID, Role: u.Role})
	return Session{Token: token, User: u}, err
}

func (s *Service) DemoUsers(ctx context.Context) ([]domain.User, error) {
	if !s.demoMode {
		return nil, domain.ErrNotFound
	}
	return s.users.DemoUsers(ctx)
}

func (s *Service) Me(ctx context.Context, p domain.Principal) (domain.User, error) {
	return s.users.ByID(ctx, p.UserID)
}
