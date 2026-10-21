package port

import "smartcredit/internal/domain"

type TokenIssuer interface {
	Issue(p domain.Principal) (string, error)
	Parse(token string) (domain.Principal, error)
}

type PasswordChecker interface {
	Check(hash, password string) bool
}
