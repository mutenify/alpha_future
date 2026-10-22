# ml-mock

Внутренний FastAPI-сервис с детерминированными моками ЕГРЮЛ, ФНС, БКИ и АБС.

Запуск из корня репозитория:

```bash
docker compose up --build ml-mock
```

В Docker-сети сервис доступен по адресу `http://ml-mock:8081`. Порт не
публикуется на хост. OpenAPI находится на `/docs`, healthcheck — на `/health`.

LLM-ручки:

- `POST /llm/recommend` — рекомендация продукта;
- `POST /llm/explain` — объяснение решения скоринга.

Без настройки модели обе ручки сразу отвечают локальными правилами с
`source: "RULES"`. Для Ollama скопируйте `.env.example` в `.env` и задайте
`OLLAMA_URL`/`OLLAMA_MODEL`. Для OpenAI-compatible API задайте `LLM_API_URL`
(полный URL `/chat/completions`), `LLM_API_KEY` и `LLM_MODEL`. Любая ошибка,
невалидный ответ или превышение 2,5 секунды автоматически включает правила.

Демо-сценарии:

- `7701234560` — ООО «Колос и Ко»;
- `5251234564` — `stopFactor: LIQUIDATION`;
- `7700000503` — `/fns/{inn}` всегда отвечает HTTP 503;
- любой другой 10- или 12-значный ИНН получает стабильные данные по хэшу.

Локальные тесты:

```bash
cd ml
python -m pip install -r requirements-dev.txt
python -m pytest
```
