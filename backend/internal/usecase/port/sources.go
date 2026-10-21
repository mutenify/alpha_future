package port

import (
	"context"
	"time"
)

// Внешние источники (Mock Gateway, api/mocks-openapi.yaml).

type EgrulData struct {
	INN, OGRN, Name, LegalForm, Status, OKVED, DirectorName string
	RegisteredAt                                            time.Time
}

type FnsData struct {
	MonthlyRevenue  []MonthAmount
	TaxDebt         float64
	AccountsBlocked bool
}

type MonthAmount struct {
	Month  string
	Amount float64
}

type BkiData struct {
	Overdues12m     int
	MaxOverdueDays  int
	Inquiries3m     int
	MonthlyPayments float64 // D_old — сумма monthlyPayment по обязательствам
}

type AbsAccount struct {
	Number  string
	Balance float64
}

type Sources interface {
	Egrul(ctx context.Context, inn string) (EgrulData, error)
	Fns(ctx context.Context, inn string) (FnsData, error)
	Bki(ctx context.Context, inn string) (BkiData, error)
	Accounts(ctx context.Context, inn string) ([]AbsAccount, error)
	Disburse(ctx context.Context, contractNumber, account string, amount float64) (time.Time, error)
}
