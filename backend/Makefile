-include .env
export

.PHONY: run build test vet migrate

run:
	go run ./cmd/api

build:
	CGO_ENABLED=0 go build -o bin/api ./cmd/api

test:
	go test ./...

vet:
	go vet ./...

# Миграции лежат в ./migrations (golang-migrate); в docker compose их применяет сервис migrate
migrate:
	migrate -path ./migrations -database "$(DATABASE_URL)" up
