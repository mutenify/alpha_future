package app

import (
	"smartcredit/internal/adapter/auth"
	"smartcredit/internal/adapter/events"
	"smartcredit/internal/adapter/explain"
	"smartcredit/internal/adapter/mocks"
	"smartcredit/internal/adapter/postgres"
	"smartcredit/internal/adapter/storage"
	"smartcredit/internal/config"
)

// infra — все адаптеры внешнего мира.
type infra struct {
	db        *postgres.DB
	repos     repos
	tokens    *auth.JWT
	passwords auth.Bcrypt
	sources   *mocks.Client
	storage   *storage.Local
	events    *events.Broker
	explainer *explain.Client
}

type repos struct {
	users        postgres.Users
	companies    postgres.Companies
	products     postgres.Products
	applications postgres.Applications
	history      postgres.History
	infos        postgres.InfoRequests
	documents    postgres.Documents
	decisions    postgres.Decisions
	offers       postgres.Offers
	contracts    postgres.Contracts
	scoring      postgres.Scoring
	rules        postgres.Rules
	queue        postgres.Queue
	audit        postgres.Audit
	jobs         postgres.Jobs
	idempotency  postgres.Idempotency
}

func newInfra(db *postgres.DB, cfg config.Config) infra {
	return infra{
		db: db,
		repos: repos{
			users: postgres.Users{DB: db}, companies: postgres.Companies{DB: db}, products: postgres.Products{DB: db},
			applications: postgres.Applications{DB: db}, history: postgres.History{DB: db},
			infos: postgres.InfoRequests{DB: db}, documents: postgres.Documents{DB: db},
			decisions: postgres.Decisions{DB: db}, offers: postgres.Offers{DB: db}, contracts: postgres.Contracts{DB: db},
			scoring: postgres.Scoring{DB: db}, rules: postgres.Rules{DB: db}, queue: postgres.Queue{DB: db},
			audit: postgres.Audit{DB: db}, jobs: postgres.Jobs{DB: db}, idempotency: postgres.Idempotency{DB: db},
		},
		tokens:    auth.NewJWT(cfg.JWTSecret, cfg.JWTTTL),
		sources:   mocks.New(cfg.MockURL, cfg.MockTimeout),
		storage:   storage.NewLocal(cfg.StorageDir),
		events:    events.NewBroker(),
		explainer: explain.New(cfg.MockURL, cfg.MockTimeout),
	}
}
