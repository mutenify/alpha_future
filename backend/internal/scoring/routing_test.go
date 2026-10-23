package scoring

import (
	"testing"

	"smartcredit/internal/domain"
)

func TestRoute(t *testing.T) {
	pr := domain.Product{PDAutoApproveMax: 0.10, PDRejectMin: 0.25}
	cases := []struct {
		pd   float64
		fits bool
		want domain.Outcome
	}{
		{0.05, true, domain.OutcomeApproved},
		{0.05, false, domain.OutcomeCounterOffer},
		{0.10, true, domain.OutcomeManualReview},
		{0.25, true, domain.OutcomeManualReview},
		{0.26, true, domain.OutcomeRejected},
	}
	for _, c := range cases {
		if got := Route(c.pd, pr, c.fits); got != c.want {
			t.Errorf("pd=%v fits=%v: got %s, want %s", c.pd, c.fits, got, c.want)
		}
	}
}
