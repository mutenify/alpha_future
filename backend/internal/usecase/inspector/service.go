// Package inspector — кабинет сотрудника: очередь, карточка заявки, решение.
package inspector

import (
	"smartcredit/internal/usecase/port"
	"smartcredit/internal/usecase/workflow"
)

type Deps struct {
	Tx        port.TxManager
	Apps      port.ApplicationRepository
	QueueRepo port.QueueRepository
	Scores    port.ScoringRepository
	Products  port.ProductRepository
	Decisions port.DecisionRepository
	Offers    port.OfferRepository
	Infos     port.InfoRequestRepository
	Flow      *workflow.Transitioner
}

type Service struct{ Deps }

func New(d Deps) *Service { return &Service{Deps: d} }
