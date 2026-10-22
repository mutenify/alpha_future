package domain

// transitions повторяет docs/process/state.mmd. Любой переход вне таблицы — ErrInvalidTransition.
var transitions = map[Status][]Status{
	StatusDraft:         {StatusSubmitted, StatusCancelled},
	StatusSubmitted:     {StatusInfoRequested, StatusEnriching},
	StatusEnriching:     {StatusRejected, StatusManualReview, StatusScoring},
	StatusScoring:       {StatusApproved, StatusCounterOffer, StatusManualReview, StatusRejected},
	StatusManualReview:  {StatusApproved, StatusCounterOffer, StatusInfoRequested, StatusRejected},
	StatusInfoRequested: {StatusSubmitted, StatusManualReview, StatusCancelled},
	StatusCounterOffer:  {StatusApproved, StatusCancelled},
	StatusApproved:      {StatusSigned, StatusCancelled},
	StatusSigned:        {StatusDisbursed},
}

func CanTransition(from, to Status) bool {
	for _, s := range transitions[from] {
		if s == to {
			return true
		}
	}
	return false
}
