package domain

type User struct {
	ID           string
	Login        string
	PasswordHash string
	Role         Role
	FullName     string
	Position     string
	ScenarioHint string
}
