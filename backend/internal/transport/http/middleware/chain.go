// Package middleware — обёртки http.Handler.
package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

// Chain применяет middleware слева направо: Chain(h, a, b) == a(b(h)).
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
