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

# Миграции лежат в ../db/migrations (golang-migrate)
migrate:
	migrate -path ../db/migrations -database "$(DATABASE_URL)" up
