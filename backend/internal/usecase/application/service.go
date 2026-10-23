// Package application — сценарии клиента над заявкой: черновик, подача, документы,
// контрпредложение, подписание.
package application

import (
	"smartcredit/internal/usecase/port"
	"smartcredit/internal/usecase/workflow"
)

type Deps struct {
	Tx        port.TxManager
	Apps      port.ApplicationRepository
	Companies port.CompanyRepository
	Products  port.ProductRepository
	History   port.HistoryRepository
	Infos     port.InfoRequestRepository
	Docs      port.DocumentRepository
	Offers    port.OfferRepository
	Decisions port.DecisionRepository
	Contracts port.ContractRepository
	Jobs      port.JobRepository
	Storage   port.FileStorage
	Sources   port.Sources
	Flow      *workflow.Transitioner
}

type Service struct{ Deps }

func New(d Deps) *Service { return &Service{Deps: d} }
