// Package app — composition root: создаёт адаптеры, use case'ы и запускает HTTP и воркер.
package app

import (
	"context"
	"log/slog"

	"smartcredit/internal/adapter/postgres"
	"smartcredit/internal/config"
	transport "smartcredit/internal/transport/http"
)

func Run(ctx context.Context, cfg config.Config) error {
	db, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	infra := newInfra(db, cfg)
	svc := newServices(infra, cfg)

	if cfg.WorkerEnable {
		go svc.pipeline.Run(ctx)
	}
	router := transport.NewRouter(transport.RouterDeps{
		Handler:     newHandler(svc, infra),
		Tokens:      infra.tokens,
		Idempotency: infra.repos.idempotency,
		CORSOrigin:  cfg.CORSOrigin,
	})
	slog.Info("core api started", "addr", cfg.HTTPAddr)
	return transport.Serve(ctx, cfg.HTTPAddr, router)
}
