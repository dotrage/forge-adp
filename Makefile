.PHONY: setup build test migrate run-local clean

setup:
	go mod download
	cd pkg/agents && poetry install

build:
	go build ./...
	cd pkg/agents && poetry build

migrate:
	@echo "Running DB migrations..."
	psql "$$DATABASE_URL" -f internal/db/migrations/000001_init_schema.up.sql

run-local:
	docker-compose -f docker-compose.dev.yml up -d
	go run ./cmd/orchestrator &
	go run ./cmd/registry &
	go run ./cmd/policy-engine &
	go run ./cmd/adapters/jira &
	go run ./cmd/adapters/github &
	go run ./cmd/adapters/slack &

test:
	go test ./...
	cd pkg/agents && poetry run pytest tests/ -v

test-integration:
	cd pkg/agents && poetry run pytest tests/integration/ -v

lint:
	go vet ./...
	cd pkg/agents && poetry run ruff check .

docker-build:
	docker build -f Dockerfile.orchestrator -t forge/orchestrator:v0.1.0 .
	docker build -f Dockerfile.agents -t forge/agents:v0.1.0 .

clean:
	go clean ./...
	find . -name '__pycache__' -exec rm -rf {} + 2>/dev/null || true
	find . -name '*.pyc' -delete 2>/dev/null || true
