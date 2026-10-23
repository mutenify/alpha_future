// Package http — сборка роутера: какой путь, какой хендлер, какие роли.
// Пути и роли соответствуют api/openapi.yaml.
package http

import (
	"net/http"

	"smartcredit/internal/transport/http/handler"
	mw "smartcredit/internal/transport/http/middleware"
	"smartcredit/internal/usecase/port"
)

const prefix = "/api/v1"

type RouterDeps struct {
	Handler     *handler.Handler
	Tokens      port.TokenIssuer
	Idempotency port.IdempotencyRepository
	CORSOrigin  string
}

func NewRouter(d RouterDeps) http.Handler {
	h := d.Handler
	mux := http.NewServeMux()
	r := routes{mux: mux, auth: mw.Auth(d.Tokens), idem: mw.Idempotency(d.Idempotency)}

	// Без токена
	r.public("GET /health", h.Health)
	r.public("GET /ready", h.Readiness)
	r.public("POST /auth/login", h.Login)
	r.public("GET /auth/demo-users", h.DemoUsers)

	// Любая роль (клиент видит только свои заявки — проверка в use case)
	r.any("GET /auth/me", h.Me)
	r.any("GET /products", h.ListProducts)
	r.any("GET /applications/{id}", h.GetApplication)
	r.any("GET /applications/{id}/history", h.History)
	r.any("GET /applications/{id}/documents", h.ListDocuments)
	r.any("GET /applications/{id}/events", h.StreamEvents)

	// Кабинет клиента
	r.client("GET /client/dashboard", h.ClientDashboard)
	r.client("POST /products/calculate", h.Calculate)
	r.client("POST /products/recommend", h.Recommend)
	r.client("GET /companies/lookup", h.LookupCompany)
	r.client("GET /applications", h.ListApplications)
	r.client("POST /applications", h.CreateApplication, r.idem)
	r.client("PATCH /applications/{id}", h.UpdateDraft)
	r.client("POST /applications/{id}/submit", h.Submit, r.idem)
	r.client("POST /applications/{id}/cancel", h.Cancel)
	r.client("POST /applications/{id}/documents", h.UploadDocument)
	r.client("POST /applications/{id}/offer/accept", h.AcceptOffer)
	r.client("POST /applications/{id}/offer/decline", h.DeclineOffer)
	r.client("POST /applications/{id}/sign", h.Sign, r.idem)

	// Кабинет инспектора
	r.inspector("GET /inspector/queue", h.Queue)
	r.inspector("GET /applications/{id}/scoring", h.Scoring)
	r.inspector("POST /applications/{id}/decision", h.Decide, r.idem)
	r.inspector("GET /rules", h.GetRules)
	r.inspector("GET /audit-log", h.AuditLog)

	return mw.Chain(mux, mw.Recover, mw.Logging, mw.CORS(d.CORSOrigin))
}
