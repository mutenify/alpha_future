package domain

import "time"

type Application struct {
	ID                  string
	Number              int64
	ClientUserID        string
	Company             Company
	Product             Product
	Amount              float64
	TermMonths          int
	Purpose             string
	Status              Status
	AssignedInspectorID string
	Version             int // оптимистичная блокировка
	CreatedAt           time.Time
	SubmittedAt         *time.Time
	UpdatedAt           time.Time
}

// ApplicationInput — поля анкеты, которые клиент заполняет в черновике.
type ApplicationInput struct {
	CompanyINN  string
	ProductCode string
	Amount      float64
	TermMonths  int
	Purpose     string
}

// ApplicationView — заявка вместе с тем, что показывает вкладка «Моя заявка».
type ApplicationView struct {
	Application
	NextStep        string
	DecisionReason  string
	Offers          []Offer
	OpenInfoRequest *InfoRequest
}

func (a Application) OwnedBy(p Principal) bool {
	return p.Role != RoleClient || a.ClientUserID == p.UserID
}
