package domain

import "testing"

func TestCanTransition(t *testing.T) {
	cases := []struct {
		from, to Status
		ok       bool
	}{
		{StatusDraft, StatusSubmitted, true},
		{StatusScoring, StatusCounterOffer, true},
		{StatusInfoRequested, StatusManualReview, true},
		{StatusDraft, StatusApproved, false},
		{StatusRejected, StatusSubmitted, false},
		{StatusSigned, StatusCancelled, false},
	}
	for _, c := range cases {
		if got := CanTransition(c.from, c.to); got != c.ok {
			t.Errorf("%s → %s: got %v, want %v", c.from, c.to, got, c.ok)
		}
	}
}
