package middleware

import (
	"net/http"
	"strings"

	"smartcredit/internal/domain"
	"smartcredit/internal/transport/http/httpx"
	"smartcredit/internal/usecase/port"
)

// Auth проверяет Bearer-токен. EventSource не умеет заголовки, поэтому для SSE
// токен можно передать в ?access_token=.
func Auth(tokens port.TokenIssuer) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if token == "" {
				token = r.URL.Query().Get("access_token")
			}
			p, err := tokens.Parse(token)
			if err != nil {
				httpx.Error(w, r, domain.ErrUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(httpx.WithPrincipal(r.Context(), p)))
		})
	}
}
