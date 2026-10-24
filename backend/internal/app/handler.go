package app

import (
	"context"

	"smartcredit/internal/transport/http/handler"
)

func newHandler(s services, in infra) *handler.Handler {
	return &handler.Handler{
		Auth: s.auth, Catalog: s.catalog, Company: s.company, Apps: s.apps, Inspector: s.inspector,
		Dashboard: s.dashboard, Rules: s.rules, Audit: s.audit, Events: in.events,
		Ready: func(ctx context.Context) map[string]string {
			return map[string]string{"db": status(in.db.Ping(ctx)), "ml": status(in.sources.Health(ctx))}
		},
	}
}

func status(err error) string {
	if err != nil {
		return "down"
	}
	return "ok"
}
