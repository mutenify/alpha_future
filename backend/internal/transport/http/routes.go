package http

import (
	"net/http"
	"strings"

	"smartcredit/internal/domain"
	mw "smartcredit/internal/transport/http/middleware"
)

// routes регистрирует "METHOD /path" под префиксом /api/v1 с нужными middleware.
type routes struct {
	mux  *http.ServeMux
	auth mw.Middleware
	idem mw.Middleware
}

func (r routes) handle(pattern string, h http.HandlerFunc, mws ...mw.Middleware) {
	method, path, _ := strings.Cut(pattern, " ")
	r.mux.Handle(method+" "+prefix+path, mw.Chain(h, mws...))
}

func (r routes) public(p string, h http.HandlerFunc) { r.handle(p, h) }

func (r routes) any(p string, h http.HandlerFunc, extra ...mw.Middleware) {
	r.handle(p, h, append([]mw.Middleware{r.auth}, extra...)...)
}

func (r routes) client(p string, h http.HandlerFunc, extra ...mw.Middleware) {
	r.handle(p, h, append([]mw.Middleware{r.auth, mw.Role(domain.RoleClient)}, extra...)...)
}

func (r routes) inspector(p string, h http.HandlerFunc, extra ...mw.Middleware) {
	r.handle(p, h, append([]mw.Middleware{r.auth, mw.Role(domain.RoleInspector, domain.RoleAdmin)}, extra...)...)
}
