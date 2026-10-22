package domain

type Outcome string

const (
	OutcomeApproved      Outcome = "APPROVED"
	OutcomeCounterOffer  Outcome = "COUNTER_OFFER"
	OutcomeManualReview  Outcome = "MANUAL_REVIEW"
	OutcomeInfoRequested Outcome = "INFO_REQUESTED"
	OutcomeRejected      Outcome = "REJECTED"
)

// Status — в какой статус переводит заявку это решение.
func (o Outcome) Status() Status { return Status(o) }

type Decision struct {
	ID                string
	ApplicationID     string
	ScoringResultID   string
	DecidedBy         ActorType // SYSTEM | INSPECTOR
	InspectorID       string
	Outcome           Outcome
	ReasonCode        string
	ReasonText        string
	ExplanationSource string // LLM | TEMPLATE
}

// InspectorDecision — ввод с кнопок карточки заявки.
type InspectorDecision struct {
	Outcome         Outcome
	ReasonText      string
	CounterOffer    *CounterTerms
	InfoRequestText string
}

type CounterTerms struct {
	ProductCode string
	Amount      float64
	TermMonths  int
}

func (d InspectorDecision) Validate() error {
	if d.ReasonText == "" {
		return Invalid("укажите причину решения")
	}
	switch d.Outcome {
	case OutcomeApproved, OutcomeRejected:
	case OutcomeCounterOffer:
		if d.CounterOffer == nil {
			return Invalid("для контрпредложения нужны сумма и срок")
		}
	case OutcomeInfoRequested:
		if d.InfoRequestText == "" {
			return Invalid("опишите, какие документы нужны")
		}
	default:
		return Invalid("неизвестное решение")
	}
	return nil
}
