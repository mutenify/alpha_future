package dto

import (
	"time"

	"smartcredit/internal/domain"
)

type ApplicationInput struct {
	CompanyINN  string  `json:"companyInn"`
	ProductCode string  `json:"productCode"`
	Amount      float64 `json:"amount"`
	TermMonths  int     `json:"termMonths"`
	Purpose     string  `json:"purpose,omitempty"`
}

func (in ApplicationInput) ToDomain() (domain.ApplicationInput, error) {
	if in.CompanyINN == "" || in.ProductCode == "" || in.Amount <= 0 || in.TermMonths <= 0 {
		return domain.ApplicationInput{}, domain.Invalid("заполните ИНН, продукт, сумму и срок")
	}
	return domain.ApplicationInput(in), nil
}

type Application struct {
	ID              string        `json:"id"`
	Number          int64         `json:"number"`
	Status          domain.Status `json:"status"`
	ProductCode     string        `json:"productCode"`
	Amount          float64       `json:"amount"`
	TermMonths      int           `json:"termMonths"`
	Purpose         string        `json:"purpose,omitempty"`
	Company         Company       `json:"company"`
	NextStep        string        `json:"nextStep,omitempty"`
	DecisionReason  string        `json:"decisionReason,omitempty"`
	Offers          []Offer       `json:"offers,omitempty"`
	OpenInfoRequest *InfoRequest  `json:"openInfoRequest,omitempty"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

func FromApplication(a domain.Application) Application {
	return Application{
		ID: a.ID, Number: a.Number, Status: a.Status, ProductCode: a.Product.Code,
		Amount: a.Amount, TermMonths: a.TermMonths, Purpose: a.Purpose,
		Company: FromCompany(a.Company), CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt,
	}
}

func FromApplications(as []domain.Application) []Application { return mapAll(as, FromApplication) }

func FromView(v domain.ApplicationView) Application {
	out := FromApplication(v.Application)
	out.NextStep, out.DecisionReason = v.NextStep, v.DecisionReason
	out.Offers = mapAll(v.Offers, FromOffer)
	if v.OpenInfoRequest != nil {
		ir := FromInfoRequest(*v.OpenInfoRequest)
		out.OpenInfoRequest = &ir
	}
	return out
}

type StatusChange struct {
	FromStatus domain.Status    `json:"fromStatus,omitempty"`
	ToStatus   domain.Status    `json:"toStatus"`
	ActorType  domain.ActorType `json:"actorType"`
	Reason     string           `json:"reason,omitempty"`
	CreatedAt  time.Time        `json:"createdAt"`
}

func FromChange(c domain.StatusChange) StatusChange {
	return StatusChange{c.From, c.To, c.ActorType, c.Reason, c.CreatedAt}
}

func FromChanges(cs []domain.StatusChange) []StatusChange { return mapAll(cs, FromChange) }
