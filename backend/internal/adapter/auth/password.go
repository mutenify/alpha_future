package auth

import "golang.org/x/crypto/bcrypt"

// Bcrypt читает хэши crypt('demo', gen_salt('bf')) из миграции 000005.
type Bcrypt struct{}

func (Bcrypt) Check(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
