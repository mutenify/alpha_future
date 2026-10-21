package domain

import "time"

type LegalForm string

const (
	LegalFormLLC LegalForm = "LLC"
	LegalFormIE  LegalForm = "IE"
)

type Company struct {
	ID                string
	OwnerUserID       string
	INN               string
	OGRN              string
	Name              string
	LegalForm         LegalForm
	OKVED             string
	RegisteredAt      *time.Time
	SettlementAccount string
}

// IndustryCode — группа ОКВЭД (первые 2 цифры) для справочника industries.
func (c Company) IndustryCode() string {
	if len(c.OKVED) < 2 {
		return "OTHER"
	}
	return c.OKVED[:2]
}
