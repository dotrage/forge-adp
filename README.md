# Forge Protocol

[![Build Status](https://img.shields.io/github/actions/workflow/status/dotrage/forge/ci.yml?branch=main&style=flat-square)](https://github.com/dotrage/forge/actions)
[![Go Version](https://img.shields.io/badge/go-1.22-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![Python Version](https://img.shields.io/badge/python-3.12-3776AB?style=flat-square&logo=python)](https://www.python.org/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue?style=flat-square)](LICENSE)

> **An enterprise agent orchestration protocol that transforms product requirements into shipped software.**

Forge connects product owners to a fleet of specialized AI agents — each responsible for a distinct software development function — through a structured workflow of planning, ticketing, execution, review, and deployment. It is designed for private deployment within any company, is not tied to a single product or tech stack, and puts human oversight at the center of every consequential action.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [API & Command Reference](#api--command-reference)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

Forge is built around four core principles:

- **Ticket-Driven Execution** — Every unit of agent work traces back to a Jira ticket. No ticket, no work.
- **Plan-Aware Context** — Agents read structured plan documents in each repository to understand the product, architecture, and constraints before acting.
- **Skills Over Prompts** — Each agent type loads a versioned, testable skill package rather than relying on ad hoc prompting.
- **Human-in-the-Loop by Default** — Agents propose. Humans approve. Every PR, schema migration, and deployment requires explicit human sign-off unless a team opts into progressive autonomy for specific low-risk action classes.

### Agent Roster

| Agent | Primary Function |
|---|---|
| **PM** | Decompose requirements into structured Jira tickets |
| **Frontend Developer** | Implement UI features from UX specs and API contracts |
| **Backend Developer** | Implement APIs, business logic, and service integrations |
| **DBA** | Design schemas, generate migrations, optimize queries |
| **DevOps** | Manage IaC, CI/CD pipelines, and deployment configs |
| **SRE** | Monitor health, define SLOs, manage incident response |
| **SecOps** | Security reviews, vulnerability assessments, compliance checks |
| **QA** | Test planning, test execution, and release validation |
| **Data Science** | Data pipelines, ML models, analytics, A/B tests |

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        FORGE CONTROL PLANE                          │
│   ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │
│   │  Orchestrator │  │   Registry   │  │     Policy Engine         │  │
│   │    Engine     │  │   Service    │  │  (permissions, autonomy,  │  │
│   └──────┬───────┘  └──────┬───────┘  │   approval workflows)    │  │
│          │                 │           └───────────┬──────────────┘  │
│   ┌──────┴─────────────────┴──────────────────────┴───────────────┐  │
│   │                    MESSAGE BUS (Event Backbone)                │  │
│   └──────┬──────────┬──────────┬──────────────────────────────────┘  │
└──────────┼──────────┼──────────┼──────────────────────────────────────┘
           │          │          │
    ┌──────┴───┐ ┌───┴────┐ ┌──┴──────┐    ...
    │  Agent   │ │ Agent  │ │  Agent  │
    │ Runtime  │ │Runtime │ │ Runtime │
    └──────────┘ └────────┘ └─────────┘
           │          │          │
    ┌──────┴──────────┴──────────┴─────────────────────────────────┐
    │                    INTEGRATION LAYER                          │
    │   Slack Adapter   |   Jira Adapter   |   GitHub Adapter      │
    └──────────────────────────────────────────────────────────────┘
```

**Control Plane** — Written in Go. Three services: Orchestrator (task routing and sequencing), Registry (agent/skill/plan catalog), Policy Engine (permissions and approval workflows).

**Agent Runtimes** — Python (3.12). Each agent runs in an isolated container with scoped LLM access, a skill loader, a plan reader, and tool access governed by the Policy Engine.

**Message Bus** — Redis Streams (default) or Apache Kafka / NATS for larger deployments.

**Storage** — PostgreSQL 16 for all persistent state (tasks, audit log, agent memory, LLM cost tracking). S3-compatible object store (MinIO locally) for skill packages and artifacts.

---

## Prerequisites

### Tools

```bash
# Go (Control Plane)
brew install go@1.22

# Python (Agent Runtimes)
brew install python@3.12
pip install poetry

# Containers and orchestration
brew install docker
brew install kubectl helm k9s

# Infrastructure
brew install terraform

# Database
brew install postgresql@16

# Message queue (local dev)
brew install redis

# Database migrations
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Code quality
brew install golangci-lint
pip install ruff black mypy
```

### API Keys and Credentials

| Service | Purpose | Environment Variable |
|---|---|---|
| Anthropic (Claude) | LLM provider | `ANTHROPIC_API_KEY` |
| GitHub | Source control | `GITHUB_TOKEN` |
| Jira (Atlassian) | Project management | `JIRA_API_TOKEN`, `JIRA_BASE_URL`, `JIRA_USER_EMAIL` |
| Slack | Communication | `SLACK_BOT_TOKEN`, `SLACK_APP_TOKEN` |
| AWS / GCP / Azure | Cloud infrastructure | Provider-specific credentials |

---

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/dotrage/forge.git
cd forge
```

### 2. Start Local Dependencies

```bash
docker-compose -f docker-compose.dev.yml up -d
```

This starts PostgreSQL 16, Redis 7, and MinIO locally.

### 3. Run Database Migrations

```bash
export DATABASE_URL="postgres://forge:forge_dev@localhost:5432/forge?sslmode=disable"
migrate -database "$DATABASE_URL" -path internal/db/migrations up
```

### 4. Build the Control Plane

```bash
make build
```

This compiles the Orchestrator, Registry, Policy Engine, and all integration adapters into the `bin/` directory.

### 5. Run Locally

```bash
make run-local
```

The Orchestrator starts on `:8080`, the Registry on `:8081`, and the Policy Engine on `:8082`.

### 6. (Optional) Seed a Project Repository

```bash
./bin/forge-seeder \
  -name "My Project" \
  -company acme \
  -project my-project \
  -output ./seeded
```

This generates a `.forge/` directory with all plan document stubs for the target repository.

### Kubernetes Deployment

```bash
# Apply namespaces
kubectl apply -f deployments/kubernetes/namespaces.yaml

# Deploy via Helm
helm install forge deployments/helm/forge \
  --namespace forge-control-plane \
  --set global.anthropicApiKey="${ANTHROPIC_API_KEY}" \
  --set global.databaseUrl="${DATABASE_URL}"
```

See `deployments/` for the full production deployment configuration including Terraform (EKS), Helm charts, and Kubernetes manifests.

---

## Usage

### Triggering Agent Work

Forge is ticket-driven. The primary way to initiate work is through Jira:

1. Create a Jira ticket with the label `forge` and assign it to the appropriate agent role.
2. The Jira Adapter polls for new tickets and emits a `task.created` event onto the message bus.
3. The Orchestrator picks up the event, resolves dependencies, and dispatches the task to the correct agent runtime.
4. The agent loads its skill, reads the project plan documents from the repository, executes, and opens a PR.
5. A `review.requested` event is emitted; the human reviewer is notified in Slack.

### Triggering via Slack

Send a message to the Forge bot in your configured Slack channel:

```
@forge implement the user authentication endpoint from ticket AUTH-42
```

The Slack Adapter parses the message, creates or links a Jira ticket, and hands off to the Orchestrator.

### Watching Activity

```bash
# Tail Orchestrator logs
kubectl logs -f -l app=forge-orchestrator -n forge-control-plane

# Watch all agent pods
kubectl get pods -n forge-agents -w
```

### Example Workflow

```
Product owner drops a PRD in Slack
    → PM Agent creates Jira epics and stories
    → Backend Developer Agent implements the API (opens PR)
    → QA Agent generates test plan and automated tests
    → SecOps Agent reviews the PR for security issues
    → Human engineer approves the PR
    → DevOps Agent updates deployment manifests
    → SRE Agent adds SLO and alert definitions
```

---

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `REDIS_URL` | Yes | `redis://localhost:6379` | Redis / message bus URL |
| `ANTHROPIC_API_KEY` | Yes | — | Anthropic Claude API key |
| `GITHUB_TOKEN` | Yes | — | GitHub personal access token |
| `JIRA_BASE_URL` | Yes | — | Jira instance URL (e.g. `https://acme.atlassian.net`) |
| `JIRA_USER_EMAIL` | Yes | — | Jira account email |
| `JIRA_API_TOKEN` | Yes | — | Jira API token |
| `SLACK_BOT_TOKEN` | Yes | — | Slack bot OAuth token |
| `SLACK_APP_TOKEN` | Yes | — | Slack app-level token (socket mode) |
| `FORGE_LOG_LEVEL` | No | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `FORGE_AUTONOMY_LEVEL` | No | `0` | Default autonomy level (0–3) for all agents |
| `FORGE_COMPANY_ID` | Yes | — | Tenant/company identifier |

### Project Repository Configuration (`.forge/config.yaml`)

Each project repository that Forge manages must contain a `.forge/` directory. The `config.yaml` controls per-project agent behavior:

```yaml
project_id: my-project
company_id: acme

agents:
  backend-developer:
    autonomy_level: 1       # Requires human PR approval
    llm_model: claude-3-5-sonnet-20241022
    skills:
      - api-implementation
      - business-logic
      - test-generation
  qa:
    autonomy_level: 1
    skills:
      - test-plan-generation
      - bug-reporting

integrations:
  jira:
    project_key: ACME
    agent_label: forge
  github:
    repo: acme-corp/my-project
    base_branch: main
  slack:
    notification_channel: "#eng-forge"
    escalation_channel: "#eng-leads"

policy:
  require_secops_review: true
  require_dba_review_for_migrations: true
  deployment_approval_required: true
```

### Plan Documents

The full plan document structure for a seeded repository:

```
.forge/
├── PRODUCT.md          # Product vision, target users, core value prop
├── ARCHITECTURE.md     # System architecture, service boundaries, tech stack
├── API_CONTRACTS.md    # API specifications (or pointer to OpenAPI files)
├── DATA_MODEL.md       # Database schema, entity relationships, data flow
├── UX_GUIDELINES.md    # Design system, component library, interaction patterns
├── SECURITY_POLICY.md  # Auth model, data classification, compliance requirements
├── INFRASTRUCTURE.md   # Cloud architecture, deployment topology, environments
├── TEST_STRATEGY.md    # Testing philosophy, coverage targets, tool choices
├── CONTRIBUTING.md     # Code standards, branch strategy, review requirements
├── GLOSSARY.md         # Domain-specific terminology
└── config.yaml         # Forge-specific configuration for this project
```

### Autonomy Levels

| Level | Behavior |
|---|---|
| `0` | Full human approval required for all agent outputs |
| `1` | Agent opens PRs; humans must approve and merge |
| `2` | Agent can merge PRs that pass all automated checks on low-risk change types |
| `3` | Agent can merge low-risk PRs (docs, tests, style) automatically after CI passes |

---

## API & Command Reference

### Control Plane HTTP API

All services expose a JSON REST API. The Orchestrator is the primary entry point.

#### Orchestrator (`:8080`)

| Method | Path | Description |
|---|---|---|
| `POST` | `/v1/tasks` | Submit a new task directly (bypasses Jira) |
| `GET` | `/v1/tasks/{id}` | Get task status and output |
| `GET` | `/v1/tasks?agent_role=&status=` | List tasks with filters |
| `POST` | `/v1/tasks/{id}/approve` | Approve a pending human-in-the-loop checkpoint |
| `POST` | `/v1/tasks/{id}/reject` | Reject a pending checkpoint with a reason |
| `GET` | `/v1/health` | Health check |

#### Registry (`:8081`)

| Method | Path | Description |
|---|---|---|
| `GET` | `/v1/agents` | List all registered agent instances |
| `GET` | `/v1/skills` | List available skills |
| `GET` | `/v1/skills/{role}/{name}` | Get a specific skill manifest |
| `POST` | `/v1/skills` | Register a new skill version |
| `GET` | `/v1/plans` | List indexed plan documents |

#### Policy Engine (`:8082`)

| Method | Path | Description |
|---|---|---|
| `POST` | `/v1/evaluate` | Evaluate a proposed agent action against policies |
| `GET` | `/v1/policies` | List active policies |
| `PUT` | `/v1/policies/{name}` | Update a policy definition |

### Makefile Commands

```bash
make setup        # Install all dependencies and tools
make migrate      # Run database migrations
make build        # Build all Go binaries to bin/
make run-local    # Start control plane services locally
make test         # Run all tests (Go + Python)
make lint         # Run golangci-lint and ruff
make docker-build # Build Docker images for all services
make seed         # Seed a local test project
```

### CLI Tools

#### Seeder

Generates `.forge/` plan documents for a new project repository:

```bash
./bin/forge-seeder \
  -name    "Project Name" \
  -company "acme" \
  -project "project-slug" \
  -output  "/path/to/repo"
```

#### Tenant Onboarding

Provisions a new enterprise tenant (namespace, database, secrets, Helm release):

```bash
./tools/tenant-onboard/main.sh <tenant-id> <tenant-name> [namespace|cluster]
# Example:
./tools/tenant-onboard/main.sh acme-corp "Acme Corporation" namespace
```

### Message Bus Events

Agents communicate through structured events on the message bus:

| Event | Description |
|---|---|
| `task.created` | New task assigned to an agent |
| `task.started` | Agent has begun execution |
| `task.blocked` | Agent cannot proceed; needs input or upstream dependency |
| `task.completed` | Agent finished; output ready for review |
| `task.failed` | Unrecoverable agent error |
| `review.requested` | Agent requests human or peer-agent review |
| `review.approved` / `review.rejected` | Review outcome |
| `deployment.requested` / `deployment.approved` | Deployment lifecycle |
| `escalation.created` | Issue surfaced to human decision-maker |

---

## Contributing

### Getting Started

1. Fork the repository and create a feature branch from `main`:
   ```bash
   git checkout -b feat/my-feature
   ```
2. Make your changes, write tests, and ensure everything passes:
   ```bash
   make test
   make lint
   ```
3. Open a pull request against `main`. All PRs require at least one reviewer approval and a passing CI build.

### Branch Strategy

| Branch prefix | Purpose |
|---|---|
| `feat/` | New features |
| `fix/` | Bug fixes |
| `chore/` | Maintenance, dependency updates |
| `docs/` | Documentation only |

### Running Tests

```bash
# Go unit + integration tests
go test ./...

# Python agent runtime tests
cd pkg/agents && poetry run pytest tests/ -v

# Load tests (requires a running local stack)
cd tests/load && k6 run orchestrator.js
```

### Code Standards

- **Go** — Follow `gofmt` and `golangci-lint` defaults. Use standard library error wrapping (`fmt.Errorf("...: %w", err)`).
- **Python** — `ruff` for linting, `black` for formatting, `mypy` for type checking (strict mode).
- **Skills** — All new skills must include a `MANIFEST.yaml`, `SKILL.md`, at least one example, and a validator.
- **Commits** — Use [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `docs:`, etc.).

### Adding a New Agent Skill

1. Create the skill directory under `skills/{agent-role}/{skill-name}/`.
2. Add `MANIFEST.yaml` with the skill metadata and autonomy level.
3. Write `SKILL.md` with execution steps, quality gates, and escalation triggers.
4. Add example input/output pairs under `examples/`.
5. Write a skill validator under `validators/`.
6. Register the skill via the Registry API or `make seed`.

See [skills/backend-developer/api-implementation/](skills/backend-developer/api-implementation/) for a reference implementation.

### Submitting a PR

- Link the relevant Jira ticket in the PR description.
- Describe what changed and why.
- Include test output showing the change works.
- Schema migrations require DBA lead + backend lead sign-off before merge.
- Any change touching authentication, secrets handling, or the Policy Engine requires a SecOps review.

---

## License

This project is licensed under the **Apache License 2.0**. See [LICENSE](LICENSE) for the full text.

---

<p align="center">
  Built with ❤️ for engineering teams who want their AI to do the work, not just talk about it.
</p>
