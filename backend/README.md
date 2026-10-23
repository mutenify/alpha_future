# Core API (Go)

Чистая архитектура: зависимости направлены только внутрь.

```
transport/http ──► usecase ──► domain
                     │  ▲
                     ▼  │
                   port ◄── adapter (postgres, mocks, auth, storage, events, explain)
                     
scoring — чистая математика, её зовут usecase'ы; app — сборка всего вместе.
```

| Слой | Пакет | Что лежит | Чего там нет |
|---|---|---|---|
| Домен | `internal/domain` | Сущности, статусы, таблица переходов (`transitions.go` = `docs/process/state.mmd`), доменные ошибки | SQL, JSON, HTTP |
| Математика | `internal/scoring` | Аннуитет, балльная карта → PD, P_gap, ставка r*, точки А/Б, маршрутизация по PD | Ввод-вывод |
| Сценарии | `internal/usecase/*` | Один пакет на кабинет/процесс, один файл на действие | Знание о Postgres и HTTP |
| Порты | `internal/usecase/port` | Интерфейсы репозиториев и внешних сервисов | Реализации |
| Смена статуса | `internal/usecase/workflow` | Единственное место, где меняется статус: проверка перехода, история, аудит, SSE | |
| Адаптеры | `internal/adapter/*` | Postgres (pgx), клиент Mock Gateway, JWT/bcrypt, файлы, SSE-брокер, шаблонные объяснения | Бизнес-правила |
| HTTP | `internal/transport/http` | `router.go` (все пути и роли), `handler/` (по файлу на вкладку), `dto/` (JSON по OpenAPI), `middleware/`, `httpx/` | Бизнес-логика |
| Сборка | `internal/app`, `cmd/api` | Создание зависимостей и запуск HTTP + воркера | |

## Карта «экран → эндпоинт → use case»

| Экран | Эндпоинт | Use case |
|---|---|---|
| Вход | `GET /auth/demo-users`, `POST /auth/login`, `GET /auth/me` | `usecase/auth` |
| Клиент · Главная | `GET /client/dashboard` | `usecase/dashboard` |
| Клиент · Подобрать кредит | `GET /products`, `POST /products/calculate`, `POST /products/recommend` | `usecase/catalog` |
| Клиент · Анкета | `GET /companies/lookup?inn=`, `POST /applications`, `PATCH /applications/{id}`, `POST/GET /applications/{id}/documents` | `usecase/company`, `usecase/application` |
| Клиент · Отправить | `POST /applications/{id}/submit` → конвейер | `application/submit.go` → `usecase/pipeline` |
| Клиент · Моя заявка | `GET /applications/{id}`, `/history`, `/events` (SSE), `/offer/accept`, `/offer/decline`, `/sign`, `/cancel` | `usecase/application` |
| Инспектор · Очередь | `GET /inspector/queue` | `inspector/queue.go` |
| Инспектор · Карточка | `GET /applications/{id}/scoring`, `POST /applications/{id}/decision` | `inspector/scoring.go`, `inspector/decision.go` |
| Инспектор · Продукты и правила | `GET /rules` | `usecase/rules` |
| Инспектор · Журнал | `GET /audit-log` | `usecase/audit` |

## Конвейер

`POST /submit` кладёт задачу в `pipeline_jobs`, воркер (`pipeline/worker.go`) забирает её `FOR UPDATE SKIP LOCKED`:
`COMPLETENESS_CHECK` → `ENRICH` (ЕГРЮЛ, ФНС, БКИ, АБС; стоп-фактор → REJECTED) → `SCORE` (PD → маршрут, офферы А/Б) → после подписи `DISBURSE`.
Ошибка шага — ретрай с задержкой 2^n с; если источник так и не ответил — MANUAL_REVIEW.

## Запуск

```bash
cp .env.example .env
make migrate   # миграции из ../db/migrations
make run
make test
```

## Где ждём математика (`TODO(математик)`)

- `scoring/cashgap.go` — формула P_gap (сейчас Φ((D_old + A − μ) / σ) без подушки C₀)
- `scoring/rate.go` — r* (задача 8)
- `scoring/factors.go` — знаменатель debt_load
- `usecase/pipeline/profile.go` — μ, σ и тренд выручки
- `scoring/testdata/golden_cases.csv` — эталон для `golden_test.go`
