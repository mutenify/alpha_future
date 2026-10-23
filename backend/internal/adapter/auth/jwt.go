// Package auth — JWT (HS256) и проверка bcrypt-паролей.
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"smartcredit/internal/domain"
)

type JWT struct {
	secret []byte
	ttl    time.Duration
}

func NewJWT(secret string, ttl time.Duration) *JWT { return &JWT{secret: []byte(secret), ttl: ttl} }

type claims struct {
	Role domain.Role `json:"role"`
	jwt.RegisteredClaims
}

func (j *JWT) Issue(p domain.Principal) (string, error) {
	c := claims{Role: p.Role, RegisteredClaims: jwt.RegisteredClaims{
		Subject:   p.UserID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.ttl)),
	}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(j.secret)
}

func (j *JWT) Parse(token string) (domain.Principal, error) {
	var c claims
	_, err := jwt.ParseWithClaims(token, &c, func(*jwt.Token) (any, error) { return j.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return domain.Principal{}, fmt.Errorf("%w: %v", domain.ErrUnauthorized, err)
	}
	return domain.Principal{UserID: c.Subject, Role: c.Role}, nil
}
