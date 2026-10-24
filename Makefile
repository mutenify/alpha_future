.PHONY: config up down go-test py-test py-test-docker e2e golden validate notebook test

config:
	docker compose --env-file .env.example config --quiet

up:
	docker compose up -d --build --wait

down:
	docker compose down

go-test:
	cd backend && go test ./...

py-test:
	cd ml && python -m pytest -m "not e2e"

py-test-docker:
	docker compose --profile test run --build --rm tests python -m pytest -m "not e2e"

e2e:
	docker compose --profile test run --build --rm tests python -m pytest -m e2e -v

golden:
	python ml/golden.py

validate:
	docker compose --profile test run --build --rm tests python validation.py

notebook:
	docker compose --profile test run --build --rm tests python run_validation_notebook.py

test: go-test py-test-docker golden validate notebook
