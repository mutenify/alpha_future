package dto

import "smartcredit/internal/domain"

type DecisionInput struct {
	Outcome                domain.Outcome `json:"outcome"`
	ReasonText             string         `json:"reasonText"`
	CounterOffer           *CounterTerms  `json:"counterOffer,omitempty"`
	InfoRequestDescription string         `json:"infoRequestDescription,omitempty"`
}

type CounterTerms struct {
	ProductCode string  `json:"productCode,omitempty"`
	Amount      float64 `json:"amount"`
	TermMonths  int     `json:"termMonths"`
}

func (in DecisionInput) ToDomain() domain.InspectorDecision {
	d := domain.InspectorDecision{Outcome: in.Outcome, ReasonText: in.ReasonText, InfoRequestText: in.InfoRequestDescription}
	if c := in.CounterOffer; c != nil {
		d.CounterOffer = &domain.CounterTerms{ProductCode: c.ProductCode, Amount: c.Amount, TermMonths: c.TermMonths}
	}
	return d
}
