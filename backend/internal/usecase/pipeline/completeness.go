package pipeline

import (
	"context"
	"strings"

	"smartcredit/internal/domain"
)

// checkCompleteness: SUBMITTED → ENRICHING или INFO_REQUESTED (запрос автопроверки).
func (s *Service) checkCompleteness(ctx context.Context, app domain.Application) error {
	if missing := missingFields(app); len(missing) > 0 {
		text := "Дополните анкету: " + strings.Join(missing, ", ")
		return s.move(ctx, app, domain.StatusInfoRequested, text, func(ctx context.Context) error {
			_, err := s.Infos.Create(ctx, domain.InfoRequest{ApplicationID: app.ID, Origin: domain.OriginAutoCheck, Description: text})
			return err
		})
	}
	return s.move(ctx, app, domain.StatusEnriching, "", func(ctx context.Context) error {
		return s.Jobs.Enqueue(ctx, app.ID, domain.StepEnrich)
	})
}

// missingFields — TODO: дополнить правилами продукта (например, документы для инвесткредита).
func missingFields(app domain.Application) []string {
	var m []string
	if strings.TrimSpace(app.Purpose) == "" {
		m = append(m, "цель кредита")
	}
	if app.Company.OKVED == "" {
		m = append(m, "ОКВЭД")
	}
	return m
}
