DROP VIEW IF EXISTS v_inspector_queue;

DROP INDEX IF EXISTS idx_audit_created;
ALTER TABLE audit_log DROP COLUMN IF EXISTS message, DROP COLUMN IF EXISTS actor_type;

ALTER TABLE decisions DROP COLUMN IF EXISTS explanation_source;

DROP INDEX IF EXISTS idx_scoring_app_latest;
ALTER TABLE scoring_results
    DROP CONSTRAINT IF EXISTS fk_scoring_model,
    DROP COLUMN IF EXISTS equilibrium_rate,
    DROP COLUMN IF EXISTS inspector_metrics;

ALTER TABLE financial_profiles
    DROP COLUMN IF EXISTS industry_code,
    DROP COLUMN IF EXISTS revenue_trend,
    DROP COLUMN IF EXISTS max_overdue_days,
    DROP COLUMN IF EXISTS stop_factor;

DROP TABLE IF EXISTS scorecard_bins, scorecard_models, industries;

ALTER TABLE companies DROP COLUMN IF EXISTS settlement_account;
ALTER TABLE users DROP COLUMN IF EXISTS position, DROP COLUMN IF EXISTS scenario_hint;
