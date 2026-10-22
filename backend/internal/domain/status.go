package domain

type Status string

const (
	StatusDraft         Status = "DRAFT"
	StatusSubmitted     Status = "SUBMITTED"
	StatusInfoRequested Status = "INFO_REQUESTED"
	StatusEnriching     Status = "ENRICHING"
	StatusScoring       Status = "SCORING"
	StatusManualReview  Status = "MANUAL_REVIEW"
	StatusCounterOffer  Status = "COUNTER_OFFER"
	StatusApproved      Status = "APPROVED"
	StatusSigned        Status = "SIGNED"
	StatusDisbursed     Status = "DISBURSED"
	StatusRejected      Status = "REJECTED"
	StatusCancelled     Status = "CANCELLED"
)

func (s Status) IsTerminal() bool {
	return s == StatusRejected || s == StatusCancelled || s == StatusDisbursed
}
