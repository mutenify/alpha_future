package dto

import (
	"time"

	"smartcredit/internal/domain"
)

type Offer struct {
	ID             string              `json:"id"`
	Kind           domain.OfferKind    `json:"kind"`
	Variant        domain.OfferVariant `json:"variant"`
	ProductCode    string              `json:"productCode"`
	Amount         float64             `json:"amount"`
	TermMonths     int                 `json:"termMonths"`
	Rate           float64             `json:"rate"`
	MonthlyPayment float64             `json:"monthlyPayment"`
	PGap           *float64            `json:"pGap,omitempty"`
	Status         domain.OfferStatus  `json:"status"`
	ExpiresAt      *time.Time          `json:"expiresAt,omitempty"`
}

func FromOffer(o domain.Offer) Offer {
	return Offer{o.ID, o.Kind, o.Variant, o.ProductCode, o.Amount, o.TermMonths, o.Rate, o.MonthlyPayment, o.PGap, o.Status, o.ExpiresAt}
}

type AcceptOfferRequest struct {
	OfferID string `json:"offerId"`
}

type SignRequest struct {
	ConfirmationCode string `json:"confirmationCode"`
}

type Contract struct {
	ContractNumber    string        `json:"contractNumber"`
	SignedAt          *time.Time    `json:"signedAt"`
	DisbursedAt       *time.Time    `json:"disbursedAt,omitempty"`
	ApplicationStatus domain.Status `json:"applicationStatus"`
}

func FromContract(c domain.Contract, s domain.Status) Contract {
	return Contract{c.ContractNumber, c.SignedAt, c.DisbursedAt, s}
}
