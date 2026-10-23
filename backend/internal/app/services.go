package app

import (
	"smartcredit/internal/config"
	"smartcredit/internal/usecase/application"
	"smartcredit/internal/usecase/audit"
	"smartcredit/internal/usecase/auth"
	"smartcredit/internal/usecase/catalog"
	"smartcredit/internal/usecase/company"
	"smartcredit/internal/usecase/dashboard"
	"smartcredit/internal/usecase/inspector"
	"smartcredit/internal/usecase/pipeline"
	"smartcredit/internal/usecase/rules"
	"smartcredit/internal/usecase/workflow"
)

type services struct {
	auth      *auth.Service
	catalog   *catalog.Service
	company   *company.Service
	apps      *application.Service
	inspector *inspector.Service
	dashboard *dashboard.Service
	rules     *rules.Service
	audit     *audit.Service
	pipeline  *pipeline.Service
}

func newServices(in infra, cfg config.Config) services {
	r := in.repos
	flow := workflow.New(r.applications, r.history, r.audit, in.events)
	apps := application.New(application.Deps{
		Tx: in.db, Apps: r.applications, Companies: r.companies, Products: r.products, History: r.history,
		Infos: r.infos, Docs: r.documents, Offers: r.offers, Decisions: r.decisions, Contracts: r.contracts,
		Jobs: r.jobs, Storage: in.storage, Sources: in.sources, Flow: flow,
	})
	return services{
		auth:    auth.New(r.users, in.tokens, in.passwords, cfg.DemoAuth),
		catalog: catalog.New(r.products),
		company: company.New(in.sources),
		apps:    apps,
		inspector: inspector.New(inspector.Deps{
			Tx: in.db, Apps: r.applications, QueueRepo: r.queue, Scores: r.scoring, Products: r.products,
			Decisions: r.decisions, Offers: r.offers, Infos: r.infos, Flow: flow,
		}),
		dashboard: dashboard.New(r.companies, r.contracts, r.history, in.sources, apps),
		rules:     rules.New(r.products, r.rules),
		audit:     audit.New(r.audit),
		pipeline: pipeline.New(pipeline.Deps{
			Tx: in.db, Jobs: r.jobs, Apps: r.applications, Infos: r.infos, Scores: r.scoring, Rules: r.rules,
			Decisions: r.decisions, Offers: r.offers, Contracts: r.contracts, Sources: in.sources,
			Explainer: in.explainer, Flow: flow,
		}),
	}
}
