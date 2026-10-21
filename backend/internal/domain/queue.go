package domain

import "time"

// QueueItem — строка v_inspector_queue.
type QueueItem struct {
	ID            string
	DisplayNumber string
	Status        Status
	CompanyName   string
	INN           string
	ProductCode   string
	ProductName   string
	Amount        float64
	TermMonths    int
	Score         *int
	ScoreDisplay  *int
	PD            *float64
	LastDecidedBy string
	LastOutcome   string
	CreatedAt     time.Time
}

type QueueCounters struct {
	ManualReview     int
	InfoRequested    int
	AutoDecidedToday int
	Total            int
}

type QueueFilter struct {
	Status Status
	Query  string
}
