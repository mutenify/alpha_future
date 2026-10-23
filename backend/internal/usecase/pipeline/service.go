// Package pipeline — асинхронный конвейер заявки на очереди pipeline_jobs:
// проверка полноты → обогащение → скоринг и решение; после подписи — выдача.
package pipeline

import (
	"smartcredit/internal/usecase/port"
	"smartcredit/internal/usecase/workflow"
)

type Deps struct {
	Tx        port.TxManager
	Jobs      port.JobRepository
	Apps      port.ApplicationRepository
	Infos     port.InfoRequestRepository
	Scores    port.ScoringRepository
	Rules     port.RulesRepository
	Decisions port.DecisionRepository
	Offers    port.OfferRepository
	Contracts port.ContractRepository
	Sources   port.Sources
	Explainer port.Explainer
	Flow      *workflow.Transitioner
}

type Service struct{ Deps }

func New(d Deps) *Service { return &Service{Deps: d} }
