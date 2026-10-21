package domain

type JobStep string

const (
	StepCompleteness JobStep = "COMPLETENESS_CHECK"
	StepEnrich       JobStep = "ENRICH"
	StepScore        JobStep = "SCORE"
	StepDisburse     JobStep = "DISBURSE"
)

type Job struct {
	ID            int64
	ApplicationID string
	Step          JobStep
	Attempts      int
	MaxAttempts   int
}
