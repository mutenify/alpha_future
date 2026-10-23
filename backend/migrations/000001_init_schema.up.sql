-- 000001: базовая схема «Умного кредитного кабинета»
-- Перечисления хранятся как text + CHECK: их проще расширять следующими миграциями, чем PG ENUM.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Пользователи и роли
CREATE TABLE users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    login         text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    role          text NOT NULL CHECK (role IN ('CLIENT', 'INSPECTOR', 'ADMIN')),
    full_name     text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now()
);

-- Компании клиента (ЮЛ или ИП)
CREATE TABLE companies (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id  uuid NOT NULL REFERENCES users (id),
    inn            varchar(12) NOT NULL UNIQUE CHECK (inn ~ '^[0-9]{10}([0-9]{2})?$'),
    ogrn           varchar(15),
    name           text NOT NULL,
    legal_form     text NOT NULL CHECK (legal_form IN ('LLC', 'IE')),
    okved          varchar(8),
    registered_at  date,
    created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_companies_owner ON companies (owner_user_id);

-- Кредитные продукты и их политика (решение отделено от скоринга)
CREATE TABLE credit_products (
    id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                 text NOT NULL UNIQUE,
    name                 text NOT NULL,
    description          text,
    min_amount           numeric(15, 2) NOT NULL CHECK (min_amount > 0),
    max_amount           numeric(15, 2) NOT NULL,
    min_term_months      int NOT NULL CHECK (min_term_months > 0),
    max_term_months      int NOT NULL,
    base_rate            numeric(5, 2) NOT NULL,  -- r₀, годовых, %
    lgd                  numeric(5, 4) NOT NULL CHECK (lgd BETWEEN 0 AND 1),
    pd_auto_approve_max  numeric(5, 4) NOT NULL,  -- порог PD клиента для автоодобрения (10%)
    pd_reject_min        numeric(5, 4) NOT NULL,  -- порог PD клиента для отказа (25%)
    gap_threshold_q      numeric(5, 4) NOT NULL CHECK (gap_threshold_q BETWEEN 0 AND 1),  -- допустимая P_gap
    payment_share_k      numeric(5, 4) NOT NULL CHECK (payment_share_k BETWEEN 0 AND 1),  -- доля потока на платежи
    is_active            boolean NOT NULL DEFAULT true,
    CHECK (max_amount >= min_amount),
    CHECK (max_term_months >= min_term_months),
    CHECK (pd_auto_approve_max < pd_reject_min)
);

-- Заявка
CREATE TABLE applications (
    id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    number                 bigint GENERATED ALWAYS AS IDENTITY UNIQUE,
    client_user_id         uuid NOT NULL REFERENCES users (id),
    company_id             uuid NOT NULL REFERENCES companies (id),
    product_id             uuid NOT NULL REFERENCES credit_products (id),
    amount                 numeric(15, 2) NOT NULL CHECK (amount > 0),
    term_months            int NOT NULL CHECK (term_months > 0),
    purpose                text,
    status                 text NOT NULL DEFAULT 'DRAFT' CHECK (status IN (
                               'DRAFT', 'SUBMITTED', 'INFO_REQUESTED', 'ENRICHING', 'SCORING',
                               'MANUAL_REVIEW', 'COUNTER_OFFER', 'APPROVED', 'SIGNED',
                               'DISBURSED', 'REJECTED', 'CANCELLED')),
    assigned_inspector_id  uuid REFERENCES users (id),
    version                int NOT NULL DEFAULT 0,  -- оптимистичная блокировка при смене статуса
    created_at             timestamptz NOT NULL DEFAULT now(),
    submitted_at           timestamptz,
    updated_at             timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_applications_client ON applications (client_user_id);
CREATE INDEX idx_applications_status ON applications (status);

-- История статусов: источник правды для таймлайна в ЛК и аудита
CREATE TABLE application_status_history (
    id              bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    application_id  uuid NOT NULL REFERENCES applications (id) ON DELETE CASCADE,
    from_status     text,
    to_status       text NOT NULL,
    actor_type      text NOT NULL CHECK (actor_type IN ('SYSTEM', 'CLIENT', 'INSPECTOR')),
    actor_user_id   uuid REFERENCES users (id),
    reason          text,
    created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_status_history_app ON application_status_history (application_id, created_at);

-- Результаты обогащения из внешних источников (моки)
CREATE TABLE enrichment_results (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id  uuid NOT NULL REFERENCES applications (id) ON DELETE CASCADE,
    source          text NOT NULL CHECK (source IN ('EGRUL', 'FNS', 'BKI')),
    status          text NOT NULL CHECK (status IN ('OK', 'ERROR', 'TIMEOUT')),
    attempt         int NOT NULL DEFAULT 1,
    payload         jsonb,
    error_message   text,
    requested_at    timestamptz NOT NULL DEFAULT now(),
    received_at     timestamptz
);
CREATE INDEX idx_enrichment_app_source ON enrichment_results (application_id, source);

-- Финансовый профиль на момент рассмотрения: нормализованные входы скоринга и модели разрыва
CREATE TABLE financial_profiles (
    id                     uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id         uuid NOT NULL REFERENCES applications (id) ON DELETE CASCADE,
    monthly_revenue        jsonb NOT NULL,  -- [{month: "2026-01", amount}] за 12 месяцев, ФНС
    revenue_cv             numeric(6, 4),   -- коэффициент вариации выручки
    cash_balance           numeric(15, 2) NOT NULL,  -- C₀, остаток на счетах
    current_debt_payments  numeric(15, 2) NOT NULL DEFAULT 0,  -- D_old, платежи по действующим кредитам в месяц
    tax_debt               numeric(15, 2) NOT NULL DEFAULT 0,
    bki_overdues_12m       int NOT NULL DEFAULT 0,
    bki_inquiries_3m       int NOT NULL DEFAULT 0,
    business_age_months    int NOT NULL,
    okved                  varchar(8),
    industry_margin        numeric(5, 4),
    cash_flow_mu           numeric(15, 2) NOT NULL,  -- μ, средний свободный поток в месяц
    cash_flow_sigma        numeric(15, 2) NOT NULL CHECK (cash_flow_sigma >= 0),  -- σ
    created_at             timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_financial_profiles_app ON financial_profiles (application_id);

-- Скоринговая оценка заёмщика (слой 1: PD клиента, без суммы и срока)
CREATE TABLE scoring_results (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id  uuid NOT NULL REFERENCES applications (id) ON DELETE CASCADE,
    financial_profile_id uuid NOT NULL REFERENCES financial_profiles (id),
    model_version   text NOT NULL,
    pd              numeric(5, 4) NOT NULL CHECK (pd BETWEEN 0 AND 1),
    score           int NOT NULL,
    factors         jsonb NOT NULL,  -- [{code, value, contribution, explanation}]
    created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_scoring_app ON scoring_results (application_id);

-- Кредитные решения (автоматические и инспектора)
CREATE TABLE decisions (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id     uuid NOT NULL REFERENCES applications (id) ON DELETE CASCADE,
    scoring_result_id  uuid REFERENCES scoring_results (id),
    decided_by         text NOT NULL CHECK (decided_by IN ('SYSTEM', 'INSPECTOR')),
    inspector_id       uuid REFERENCES users (id),
    outcome            text NOT NULL CHECK (outcome IN (
                           'APPROVED', 'COUNTER_OFFER', 'MANUAL_REVIEW', 'INFO_REQUESTED', 'REJECTED')),
    reason_code        text,
    reason_text        text,  -- объяснение простым языком для клиента
    created_at         timestamptz NOT NULL DEFAULT now(),
    CHECK (decided_by = 'SYSTEM' OR inspector_id IS NOT NULL)
);
CREATE INDEX idx_decisions_app ON decisions (application_id);

-- Предложения условий (исходное и контрпредложения)
CREATE TABLE offers (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id   uuid NOT NULL REFERENCES applications (id) ON DELETE CASCADE,
    decision_id      uuid NOT NULL REFERENCES decisions (id),
    product_id       uuid NOT NULL REFERENCES credit_products (id),
    kind             text NOT NULL CHECK (kind IN ('ORIGINAL', 'COUNTER')),
    variant          text NOT NULL DEFAULT 'REQUESTED' CHECK (variant IN ('REQUESTED', 'POINT_A', 'POINT_B', 'INSPECTOR')),  -- A: меньше сумма, B: дольше срок
    amount           numeric(15, 2) NOT NULL CHECK (amount > 0),
    term_months      int NOT NULL CHECK (term_months > 0),
    rate             numeric(5, 2) NOT NULL,
    monthly_payment  numeric(15, 2) NOT NULL,
    p_gap            numeric(5, 4) CHECK (p_gap BETWEEN 0 AND 1),  -- слой 2: вероятность кассового разрыва при этих условиях
    status           text NOT NULL DEFAULT 'PROPOSED' CHECK (status IN ('PROPOSED', 'ACCEPTED', 'DECLINED', 'EXPIRED')),
    expires_at       timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_offers_app ON offers (application_id);

-- Запросы дополнительных данных; origin определяет, куда вернётся заявка
CREATE TABLE info_requests (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id  uuid NOT NULL REFERENCES applications (id) ON DELETE CASCADE,
    origin          text NOT NULL CHECK (origin IN ('AUTO_CHECK', 'INSPECTOR')),
    requested_by    uuid REFERENCES users (id),
    description     text NOT NULL,
    status          text NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'FULFILLED', 'EXPIRED', 'CANCELLED')),
    created_at      timestamptz NOT NULL DEFAULT now(),
    fulfilled_at    timestamptz,
    CHECK (origin = 'AUTO_CHECK' OR requested_by IS NOT NULL)
);
CREATE INDEX idx_info_requests_app ON info_requests (application_id);

-- Документы клиента (сами файлы в объектном хранилище)
CREATE TABLE documents (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id   uuid NOT NULL REFERENCES applications (id) ON DELETE CASCADE,
    info_request_id  uuid REFERENCES info_requests (id),
    doc_type         text NOT NULL,
    file_name        text NOT NULL,
    storage_key      text NOT NULL UNIQUE,
    mime_type        text NOT NULL,
    size_bytes       bigint NOT NULL CHECK (size_bytes > 0),
    uploaded_by      uuid NOT NULL REFERENCES users (id),
    uploaded_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_documents_app ON documents (application_id);

-- Договор, подписание (ПЭП) и выдача (смоделированы)
CREATE TABLE contracts (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id      uuid NOT NULL UNIQUE REFERENCES applications (id) ON DELETE CASCADE,
    offer_id            uuid NOT NULL REFERENCES offers (id),
    contract_number     text NOT NULL UNIQUE,
    signed_at           timestamptz,
    signature_hash      text,
    disbursed_at        timestamptz,
    settlement_account  varchar(20)
);

-- Аудит действий пользователей и системы
CREATE TABLE audit_log (
    id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    actor_user_id  uuid REFERENCES users (id),
    action         text NOT NULL,
    entity_type    text NOT NULL,
    entity_id      uuid,
    payload        jsonb,
    created_at     timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_entity ON audit_log (entity_type, entity_id);
