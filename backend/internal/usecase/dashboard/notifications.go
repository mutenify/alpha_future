package dashboard

import (
	"context"
	"fmt"

	"smartcredit/internal/domain"
)

// notifications: открытые запросы документов, предложения и последние смены статусов.
func (s *Service) notifications(ctx context.Context, p domain.Principal, views []domain.ApplicationView) ([]domain.Notification, error) {
	var out []domain.Notification
	for _, v := range views {
		if r := v.OpenInfoRequest; r != nil {
			out = append(out, domain.Notification{Kind: "INFO_REQUESTED", Text: r.Description, ApplicationID: v.ID, CreatedAt: v.UpdatedAt})
		}
		if v.Status == domain.StatusCounterOffer {
			out = append(out, domain.Notification{Kind: "OFFER", Text: "Банк предложил другие условия по заявке", ApplicationID: v.ID, CreatedAt: v.UpdatedAt})
		}
	}
	changes, err := s.history.RecentByClient(ctx, p.UserID, recentStatusChanges)
	if err != nil {
		return nil, err
	}
	for _, c := range changes {
		out = append(out, domain.Notification{
			Kind: "STATUS", Text: fmt.Sprintf("Статус заявки: %s", c.To), ApplicationID: c.ApplicationID, CreatedAt: c.CreatedAt,
		})
	}
	return out, nil
}
