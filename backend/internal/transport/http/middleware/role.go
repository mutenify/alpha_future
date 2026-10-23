package middleware

import (
	"net/http"
	"slices"

	"smartcredit/internal/domain"
	"smartcredit/internal/transport/http/httpx"
)

// Role пропускает только перечисленные роли (ставится после Auth).
func Role(roles ...domain.Role) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !slices.Contains(roles, httpx.Principal(r.Context()).Role) {
				httpx.Error(w, r, domain.ErrForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
