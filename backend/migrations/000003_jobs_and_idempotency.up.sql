-- 000003: надёжность конвейера и идемпотентность API

-- Очередь задач конвейера в Postgres: заявка не теряется при рестарте Core.
-- Воркер забирает задачу: SELECT ... WHERE status = 'PENDING' AND run_at <= now()
--   ORDER BY run_at FOR UPDATE SKIP LOCKED LIMIT 1, затем ставит RUNNING и locked_until.
-- Задача RUNNING с истёкшим locked_until считается брошенной и берётся снова.
CREATE TABLE pipeline_jobs (
    id              bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    application_id  uuid NOT NULL REFERENCES applications (id) ON DELETE CASCADE,
    step            text NOT NULL CHECK (step IN ('COMPLETENESS_CHECK', 'ENRICH', 'SCORE', 'DISBURSE')),
    status          text NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'RUNNING', 'DONE', 'FAILED')),
    attempts        int NOT NULL DEFAULT 0,
    max_attempts    int NOT NULL DEFAULT 5,
    run_at          timestamptz NOT NULL DEFAULT now(),  -- следующая попытка, экспоненциальная задержка
    locked_until    timestamptz,
    last_error      text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_pipeline_jobs_pending ON pipeline_jobs (run_at) WHERE status IN ('PENDING', 'RUNNING');
-- Не больше одной активной задачи на шаг заявки
CREATE UNIQUE INDEX uq_pipeline_jobs_active ON pipeline_jobs (application_id, step) WHERE status IN ('PENDING', 'RUNNING');

-- Идемпотентность изменяющих запросов (заголовок Idempotency-Key):
-- повтор с тем же ключом возвращает сохранённый ответ, а не создаёт вторую заявку.
CREATE TABLE idempotency_keys (
    key              text NOT NULL,
    user_id          uuid NOT NULL REFERENCES users (id),
    request_method   text NOT NULL,
    request_path     text NOT NULL,
    request_hash     text NOT NULL,  -- тот же ключ с другим телом даёт 422
    response_status  int,
    response_body    jsonb,
    created_at       timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, key)
);
CREATE INDEX idx_idempotency_created ON idempotency_keys (created_at);
