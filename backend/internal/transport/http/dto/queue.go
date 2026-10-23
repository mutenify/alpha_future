package dto

import (
	"time"

	"smartcredit/internal/domain"
)

type Queue struct {
	Counters QueueCounters `json:"counters"`
	Items    []QueueItem   `json:"items"`
}

type QueueCounters struct {
	ManualReview     int `json:"manualReview"`
	InfoRequested    int `json:"infoRequested"`
	AutoDecidedToday int `json:"autoDecidedToday"`
	Total            int `json:"total"`
}

type QueueItem struct {
	ID            string        `json:"id"`
	DisplayNumber string        `json:"displayNumber"`
	Status        domain.Status `json:"status"`
	CompanyName   string        `json:"companyName"`
	INN           string        `json:"inn"`
	ProductCode   string        `json:"productCode"`
	ProductName   string        `json:"productName,omitempty"`
	Amount        float64       `json:"amount"`
	TermMonths    int           `json:"termMonths"`
	Score         *int          `json:"score,omitempty"`
	ScoreDisplay  *int          `json:"scoreDisplay,omitempty"`
	PD            *float64      `json:"pd,omitempty"`
	LastDecidedBy string        `json:"lastDecidedBy,omitempty"`
	LastOutcome   string        `json:"lastOutcome,omitempty"`
	CreatedAt     time.Time     `json:"createdAt"`
}

func FromQueue(items []domain.QueueItem, c domain.QueueCounters) Queue {
	return Queue{
		Counters: QueueCounters(c),
		Items:    mapAll(items, func(q domain.QueueItem) QueueItem { return QueueItem(q) }),
	}
}
