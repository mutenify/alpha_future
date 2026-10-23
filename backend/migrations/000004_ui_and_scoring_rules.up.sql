-- 000004: данные для экранов кабинетов и правила скоринга как данные, а не хардкод в Go.
-- Добавляет к схеме 000001–000003 то, что нужно экранам «Очередь заявок», «Карточка заявки»,
-- «Продукты и правила», «Журнал действий» и демо-переключателю ролей.

-- Демо-переключатель ролей: подпись в карточке профиля и подсказка сценария для жюри.
-- Роль остаётся INSPECTOR (так в OpenAPI и JWT), «Андеррайтер МСБ» — это position.
ALTER TABLE users
    ADD COLUMN position       text,
    ADD COLUMN scenario_hint  text;

-- Расчётный счёт для блока «Ваши продукты» на главной клиента (из мока АБС)
ALTER TABLE companies
    ADD COLUMN settlement_account varchar(20) CHECK (settlement_account ~ '^[0-9]{20}$');

-- Справочник отраслей: параметры модели кассового разрыва по группе ОКВЭД (первые 2 цифры).
-- Значения совпадают с INDUSTRIES в math/generate_companies.py; OTHER — фолбэк для прочих ОКВЭД.
CREATE TABLE industries (
    code        text PRIMARY KEY,
    name        text NOT NULL,
    margin      numeric(5, 4) NOT NULL CHECK (margin BETWEEN -1 AND 1),  -- маржа FCF к выручке
    revenue_cv  numeric(5, 4) NOT NULL CHECK (revenue_cv >= 0)           -- типичный коэф. вариации выручки
);

INSERT INTO industries (code, name, margin, revenue_cv) VALUES
    ('10', 'Производство',        0.0900, 0.3000),
    ('41', 'Строительство',       0.1000, 0.4500),
    ('47', 'Розничная торговля',  0.0800, 0.2500),
    ('49', 'Грузоперевозки',      0.1200, 0.3000),
    ('56', 'Общепит',             0.1000, 0.3500),
    ('62', 'IT-услуги',           0.2000, 0.3000),
    ('OTHER', 'Прочие отрасли',   0.1000, 0.3500);

-- Версии балльной карты. PD = 1 / (1 + base_odds * 2^((score - base_score) / pdo)).
-- score_min/score_max — сумма минимальных и максимальных баллов по факторам,
-- по ним Go переводит балл в шкалу 0–100 для колонки «БАЛЛ» в очереди.
CREATE TABLE scorecard_models (
    version     text PRIMARY KEY,
    base_score  int NOT NULL,
    base_odds   numeric(8, 2) NOT NULL CHECK (base_odds > 0),
    pdo         int NOT NULL CHECK (pdo > 0),
    score_min   int NOT NULL,
    score_max   int NOT NULL,
    is_active   boolean NOT NULL DEFAULT false,
    created_at  timestamptz NOT NULL DEFAULT now(),
    CHECK (score_max > score_min)
);
-- Активна ровно одна версия
CREATE UNIQUE INDEX uq_scorecard_models_active ON scorecard_models (is_active) WHERE is_active;

-- Баллы карты по интервалам. Числовой бин: (lower_bound, upper_bound], NULL — бесконечность.
-- Категориальный бин (okved_group): category = код отрасли.
CREATE TABLE scorecard_bins (
    id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    model_version  text NOT NULL REFERENCES scorecard_models (version) ON DELETE CASCADE,
    factor         text NOT NULL,
    lower_bound    numeric,
    upper_bound    numeric,
    category       text,
    points         int NOT NULL,
    explanation    text,  -- фраза для объяснения решения (шаблоны математика, задача 7)
    CHECK (category IS NULL OR (lower_bound IS NULL AND upper_bound IS NULL)),
    CHECK (category IS NOT NULL OR lower_bound IS NOT NULL OR upper_bound IS NOT NULL),
    CHECK (lower_bound IS NULL OR upper_bound IS NULL OR lower_bound < upper_bound),
    UNIQUE NULLS NOT DISTINCT (model_version, factor, lower_bound, upper_bound, category)
);

-- Карта v1 из math/scorecard_points.csv
INSERT INTO scorecard_models (version, base_score, base_odds, pdo, score_min, score_max, is_active)
VALUES ('v1', 600, 50, 20, 457, 621, true);

INSERT INTO scorecard_bins (model_version, factor, lower_bound, upper_bound, category, points) VALUES
    ('v1', 'age_months', 0, 12, NULL, 46),
    ('v1', 'age_months', 12, 24, NULL, 51),
    ('v1', 'age_months', 24, 60, NULL, 70),
    ('v1', 'age_months', 60, NULL, NULL, 66),
    ('v1', 'overdues12m', -1, 0, NULL, 74),
    ('v1', 'overdues12m', 0, 1, NULL, 62),
    ('v1', 'overdues12m', 1, 3, NULL, 40),
    ('v1', 'overdues12m', 3, NULL, NULL, 50),
    ('v1', 'max_overdue_days', -1, 0, NULL, 63),
    ('v1', 'max_overdue_days', 0, 30, NULL, 62),
    ('v1', 'max_overdue_days', 30, NULL, NULL, 62),
    ('v1', 'inquiries3m', -1, 0, NULL, 64),
    ('v1', 'inquiries3m', 0, 2, NULL, 63),
    ('v1', 'inquiries3m', 2, NULL, NULL, 58),
    ('v1', 'tax_debt', -1, 0, NULL, 63),
    ('v1', 'tax_debt', 0, NULL, NULL, 62),
    ('v1', 'revenue_trend', NULL, -0.1, NULL, 61),
    ('v1', 'revenue_trend', -0.1, 0.1, NULL, 74),
    ('v1', 'revenue_trend', 0.1, NULL, NULL, 47),
    ('v1', 'revenue_cv', 0, 0.25, NULL, 63),
    ('v1', 'revenue_cv', 0.25, 0.4, NULL, 64),
    ('v1', 'revenue_cv', 0.4, NULL, NULL, 60),
    ('v1', 'debt_load', -1, 0, NULL, 72),
    ('v1', 'debt_load', 0, 0.3, NULL, 62),
    ('v1', 'debt_load', 0.3, 0.6, NULL, 58),
    ('v1', 'debt_load', 0.6, NULL, NULL, 44),
    ('v1', 'okved_group', NULL, NULL, '10', 71),
    ('v1', 'okved_group', NULL, NULL, '41', 38),
    ('v1', 'okved_group', NULL, NULL, '47', 77),
    ('v1', 'okved_group', NULL, NULL, '49', 71),
    ('v1', 'okved_group', NULL, NULL, '56', 56),
    ('v1', 'okved_group', NULL, NULL, '62', 62);

-- Снимок на момент рассмотрения: недостающие входы карты, стоп-фактор и отрасль
ALTER TABLE financial_profiles
    ADD COLUMN industry_code     text REFERENCES industries (code),
    ADD COLUMN revenue_trend     numeric(6, 4),
    ADD COLUMN max_overdue_days  int NOT NULL DEFAULT 0,
    ADD COLUMN stop_factor       text CHECK (stop_factor IN ('LIQUIDATION', 'BANKRUPTCY', 'ACCOUNTS_BLOCKED'));

-- Каждый пересчёт после INFO_REQUESTED или MANUAL_REVIEW — новая строка; актуальная — последняя
ALTER TABLE scoring_results
    ADD CONSTRAINT fk_scoring_model FOREIGN KEY (model_version) REFERENCES scorecard_models (version),
    ADD COLUMN equilibrium_rate   numeric(5, 2),  -- r*, годовых, % (задача 8)
    ADD COLUMN inspector_metrics  jsonb;          -- 5 индикаторов: [{code, value, color: GREEN|YELLOW|RED}]
CREATE INDEX idx_scoring_app_latest ON scoring_results (application_id, created_at DESC);

-- Откуда текст объяснения: LLM или шаблонный фолбэк
ALTER TABLE decisions
    ADD COLUMN explanation_source text CHECK (explanation_source IN ('LLM', 'TEMPLATE'));

-- «Журнал действий»: кто действовал и строка для ленты
ALTER TABLE audit_log
    ADD COLUMN actor_type text NOT NULL DEFAULT 'SYSTEM' CHECK (actor_type IN ('SYSTEM', 'CLIENT', 'INSPECTOR')),
    ADD COLUMN message    text;
CREATE INDEX idx_audit_created ON audit_log (created_at DESC);

-- Очередь заявок для кабинета инспектора: номер Z-…, балл, PD и кто принял последнее решение.
-- Счётчики над таблицей считаются по этому же представлению.
CREATE VIEW v_inspector_queue AS
SELECT a.id,
       'Z-' || a.number                 AS display_number,
       a.status,
       a.amount,
       a.term_months,
       a.assigned_inspector_id,
       a.created_at,
       a.updated_at,
       c.name                           AS company_name,
       c.inn,
       p.code                           AS product_code,
       p.name                           AS product_name,
       s.score,
       round(100.0 * (s.score - m.score_min) / (m.score_max - m.score_min))::int AS score_display,
       s.pd,
       d.decided_by                     AS last_decided_by,
       d.outcome                        AS last_outcome
FROM applications a
JOIN companies c       ON c.id = a.company_id
JOIN credit_products p ON p.id = a.product_id
LEFT JOIN LATERAL (
    SELECT sr.score, sr.pd, sr.model_version
    FROM scoring_results sr
    WHERE sr.application_id = a.id
    ORDER BY sr.created_at DESC
    LIMIT 1
) s ON true
LEFT JOIN scorecard_models m ON m.version = s.model_version
LEFT JOIN LATERAL (
    SELECT dc.decided_by, dc.outcome
    FROM decisions dc
    WHERE dc.application_id = a.id
    ORDER BY dc.created_at DESC
    LIMIT 1
) d ON true
WHERE a.status <> 'DRAFT';
