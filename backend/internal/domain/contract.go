package domain

import "time"

type Contract struct {
	ID                string
	ApplicationID     string
	OfferID           string
	ContractNumber    string
	SignedAt          *time.Time
	SignatureHash     string
	DisbursedAt       *time.Time
	SettlementAccount string
}
