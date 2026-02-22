.PHONY: setup build build-mcp build-vscode-ext install-vscode-ext test migrate run-local clean

setup:
	go mod download
	cd pkg/agents && poetry install
	cd tools/mcp-server && npm install
	cd tools/vscode-extension && npm install

build:
	go build ./...
	cd pkg/agents && poetry build
	$(MAKE) build-mcp
	$(MAKE) build-vscode-ext

build-mcp:
	@echo "Building Forge MCP server..."
	cd tools/mcp-server && npm run build

build-vscode-ext:
	@echo "Building Forge VS Code extension..."
	cd tools/vscode-extension && npm run build

install-vscode-ext: build-vscode-ext
	@echo "Packaging and installing VS Code extension..."
	cd tools/vscode-extension && \
		npx --yes @vscode/vsce package --no-dependencies && \
		code --install-extension forge-adp-*.vsix

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
	go run ./cmd/adapters/gitlab &
	go run ./cmd/adapters/confluence &
	go run ./cmd/adapters/linear &

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
	rm -rf tools/mcp-server/dist tools/mcp-server/node_modules
	rm -rf tools/vscode-extension/dist tools/vscode-extension/node_modules tools/vscode-extension/*.vsix
