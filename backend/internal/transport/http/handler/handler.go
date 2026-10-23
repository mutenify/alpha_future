// Package handler — HTTP-хендлеры: разбор запроса → use case → DTO.
// Бизнес-логики здесь нет.
package handler

import (
	"context"

	"smartcredit/internal/usecase/application"
	"smartcredit/internal/usecase/audit"
	"smartcredit/internal/usecase/auth"
	"smartcredit/internal/usecase/catalog"
	"smartcredit/internal/usecase/company"
	"smartcredit/internal/usecase/dashboard"
	"smartcredit/internal/usecase/inspector"
	"smartcredit/internal/usecase/port"
	"smartcredit/internal/usecase/rules"
)

type Handler struct {
	Auth      *auth.Service
	Catalog   *catalog.Service
	Company   *company.Service
	Apps      *application.Service
	Inspector *inspector.Service
	Dashboard *dashboard.Service
	Rules     *rules.Service
	Audit     *audit.Service
	Events    port.EventSubscriber
	Ready     func(ctx context.Context) map[string]string
}
