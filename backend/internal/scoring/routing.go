package scoring

import "smartcredit/internal/domain"

// Route — маршрутизация по PD клиента и порогам продукта:
// PD > 25% → отказ, 10–25% → ручной разбор, < 10% → одобрение или контрпредложение.
func Route(pd float64, pr domain.Product, fitsLimit bool) domain.Outcome {
	switch {
	case pd > pr.PDRejectMin:
		return domain.OutcomeRejected
	case pd >= pr.PDAutoApproveMax:
		return domain.OutcomeManualReview
	case fitsLimit:
		return domain.OutcomeApproved
	default:
		return domain.OutcomeCounterOffer
	}
}
