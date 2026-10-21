package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"smartcredit/internal/app"
	"smartcredit/internal/config"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := app.Run(ctx, cfg); err != nil {
		slog.Error("core api stopped", "err", err)
		os.Exit(1)
	}
}
