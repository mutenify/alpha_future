package domain

import "time"

type OfferKind string

const (
	OfferOriginal OfferKind = "ORIGINAL"
	OfferCounter  OfferKind = "COUNTER"
)

type OfferVariant string

const (
	VariantRequested OfferVariant = "REQUESTED"
	VariantPointA    OfferVariant = "POINT_A" // меньше сумма, тот же срок
	VariantPointB    OfferVariant = "POINT_B" // та же сумма, дольше срок
	VariantInspector OfferVariant = "INSPECTOR"
)

type OfferStatus string

const (
	OfferProposed OfferStatus = "PROPOSED"
	OfferAccepted OfferStatus = "ACCEPTED"
	OfferDeclined OfferStatus = "DECLINED"
	OfferExpired  OfferStatus = "EXPIRED"
)

type Offer struct {
	ID             string
	ApplicationID  string
	DecisionID     string
	ProductID      string
	ProductCode    string
	Kind           OfferKind
	Variant        OfferVariant
	Amount         float64
	TermMonths     int
	Rate           float64
	MonthlyPayment float64
	PGap           *float64
	Status         OfferStatus
	ExpiresAt      *time.Time
}
