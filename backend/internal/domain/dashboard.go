package domain

import "time"

type Dashboard struct {
	Company       Company
	Account       *Account
	ActiveLoans   []Loan
	Applications  []ApplicationView
	Notifications []Notification
}

type Account struct {
	Number  string
	Balance float64
}

type Loan struct {
	ContractNumber string
	ProductCode    string
	Amount         float64
	TermMonths     int
	Rate           float64
	MonthlyPayment float64
	DisbursedAt    *time.Time
}

type Notification struct {
	Kind          string // INFO_REQUESTED | OFFER | STATUS
	Text          string
	ApplicationID string
	CreatedAt     time.Time
}
