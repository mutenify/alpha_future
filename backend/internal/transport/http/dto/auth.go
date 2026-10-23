package dto

import "smartcredit/internal/domain"

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password,omitempty"`
}

type LoginResponse struct {
	AccessToken string      `json:"accessToken"`
	Role        domain.Role `json:"role"`
	User        UserProfile `json:"user"`
}

type UserProfile struct {
	ID           string      `json:"id"`
	Login        string      `json:"login"`
	Role         domain.Role `json:"role"`
	FullName     string      `json:"fullName"`
	Position     string      `json:"position,omitempty"`
	ScenarioHint string      `json:"scenarioHint,omitempty"`
}

func FromUser(u domain.User) UserProfile {
	return UserProfile{u.ID, u.Login, u.Role, u.FullName, u.Position, u.ScenarioHint}
}

func FromUsers(us []domain.User) []UserProfile { return mapAll(us, FromUser) }
