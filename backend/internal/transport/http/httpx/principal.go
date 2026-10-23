package httpx

import (
	"context"

	"smartcredit/internal/domain"
)

type principalKey struct{}

func WithPrincipal(ctx context.Context, p domain.Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

// Principal — пользователь из JWT; middleware.Auth гарантирует, что он есть.
func Principal(ctx context.Context) domain.Principal {
	p, _ := ctx.Value(principalKey{}).(domain.Principal)
	return p
}
