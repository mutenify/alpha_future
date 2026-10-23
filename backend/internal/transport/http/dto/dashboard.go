package dto

import (
	"time"

	"smartcredit/internal/domain"
)

type Dashboard struct {
	Company       Company        `json:"company"`
	Account       *Account       `json:"account,omitempty"`
	ActiveLoans   []Loan         `json:"activeLoans"`
	Applications  []Application  `json:"applications"`
	Notifications []Notification `json:"notifications"`
}

type Account struct {
	Number  string  `json:"number"`
	Balance float64 `json:"balance"`
}

type Loan struct {
	ContractNumber string     `json:"contractNumber"`
	ProductCode    string     `json:"productCode"`
	Amount         float64    `json:"amount"`
	TermMonths     int        `json:"termMonths"`
	Rate           float64    `json:"rate"`
	MonthlyPayment float64    `json:"monthlyPayment"`
	DisbursedAt    *time.Time `json:"disbursedAt,omitempty"`
}

type Notification struct {
	Kind          string    `json:"kind"`
	Text          string    `json:"text"`
	ApplicationID string    `json:"applicationId,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

func FromDashboard(d domain.Dashboard) Dashboard {
	out := Dashboard{
		Company:       FromCompany(d.Company),
		ActiveLoans:   mapAll(d.ActiveLoans, func(l domain.Loan) Loan { return Loan(l) }),
		Applications:  mapAll(d.Applications, FromView),
		Notifications: mapAll(d.Notifications, func(n domain.Notification) Notification { return Notification(n) }),
	}
	if d.Account != nil {
		a := Account(*d.Account)
		out.Account = &a
	}
	return out
}
