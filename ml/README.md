# ml-mock-service

Внутренний FastAPI-сервис с детерминированными моками ЕГРЮЛ, ФНС, БКИ и АБС.

Запуск из корня репозитория:

```bash
docker compose up --build ml-mock-service
```

В Docker-сети сервис доступен по адресу `http://ml-mock-service:8081`. Порт не
публикуется на хост. OpenAPI находится на `/docs`, healthcheck — на `/health`.

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
