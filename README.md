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
    │  Jira · GitHub · GitLab · Slack · Teams · Google Chat        │
    │  Confluence · Linear · PagerDuty · Opsgenie                  │
    │  Datadog · Grafana · Snyk · SonarQube                        │
    │  Terraform Cloud · Atlantis                                  │
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
| GitHub | Source control | `GITHUB_TOKEN`, `GITHUB_WEBHOOK_SECRET` |
| GitLab | Source control (self-hosted or GitLab.com) | `GITLAB_TOKEN`, `GITLAB_BASE_URL`, `GITLAB_WEBHOOK_SECRET` |
| Jira (Atlassian) | Project management | `JIRA_API_TOKEN`, `JIRA_BASE_URL`, `JIRA_USER_EMAIL` |
| Confluence (Atlassian) | Documentation & specs | `CONFLUENCE_BASE_URL`, `CONFLUENCE_USERNAME`, `CONFLUENCE_API_TOKEN` |
| Linear | Issue tracking | `LINEAR_API_KEY`, `LINEAR_WEBHOOK_SECRET` |
| Slack | Team communication | `SLACK_BOT_TOKEN`, `SLACK_APP_TOKEN` |
| Microsoft Teams | Team communication | `TEAMS_WEBHOOK_URL`, `TEAMS_HMAC_SECRET` |
| Google Chat | Team communication | `GOOGLE_CHAT_WEBHOOK_URL`, `GOOGLE_CHAT_VERIFICATION_TOKEN` |
| PagerDuty | Incident management | `PAGERDUTY_API_KEY`, `PAGERDUTY_SERVICE_ID`, `PAGERDUTY_FROM_EMAIL` |
| Opsgenie | Alert management | `OPSGENIE_API_KEY` |
| Datadog | Metrics & monitoring | `DATADOG_API_KEY`, `DATADOG_APP_KEY` |
| Grafana | Dashboards & alerting | `GRAFANA_URL`, `GRAFANA_API_KEY` |
| Snyk | Vulnerability scanning | `SNYK_API_TOKEN`, `SNYK_ORG_ID`, `SNYK_WEBHOOK_SECRET` |
| SonarQube | Static code analysis | `SONARQUBE_URL`, `SONARQUBE_TOKEN`, `SONARQUBE_WEBHOOK_SECRET` |
| Terraform Cloud | IaC plan & apply automation | `TFC_TOKEN`, `TFC_ORGANIZATION`, `TFC_WEBHOOK_HMAC_KEY` |
| Atlantis | GitOps Terraform automation | `ATLANTIS_URL`, `ATLANTIS_TOKEN`, `ATLANTIS_WEBHOOK_SECRET` |
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
| `GITHUB_WEBHOOK_SECRET` | Yes | — | Secret used to validate GitHub webhook payloads |
| `GITLAB_TOKEN` | No¹ | — | GitLab personal access token |
| `GITLAB_BASE_URL` | No¹ | `https://gitlab.com` | GitLab instance base URL (override for self-hosted) |
| `GITLAB_WEBHOOK_SECRET` | No¹ | — | Secret used to validate GitLab webhook `X-Gitlab-Token` header |
| `JIRA_BASE_URL` | Yes | — | Jira instance URL (e.g. `https://acme.atlassian.net`) |
| `JIRA_USER_EMAIL` | Yes | — | Jira account email |
| `JIRA_API_TOKEN` | Yes | — | Jira API token |
| `CONFLUENCE_BASE_URL` | No¹ | — | Confluence instance URL (e.g. `https://acme.atlassian.net`) |
| `CONFLUENCE_USERNAME` | No¹ | — | Confluence account email |
| `CONFLUENCE_API_TOKEN` | No¹ | — | Confluence API token |
| `LINEAR_API_KEY` | No¹ | — | Linear API key |
| `LINEAR_WEBHOOK_SECRET` | No¹ | — | Secret used to validate Linear webhook HMAC-SHA256 signatures |
| `SLACK_BOT_TOKEN` | No¹ | — | Slack bot OAuth token |
| `SLACK_APP_TOKEN` | No¹ | — | Slack app-level token (socket mode) |
| `TEAMS_WEBHOOK_URL` | No¹ | — | Microsoft Teams incoming webhook URL |
| `TEAMS_HMAC_SECRET` | No¹ | — | HMAC secret for validating Teams bot activity payloads |
| `GOOGLE_CHAT_WEBHOOK_URL` | No¹ | — | Google Chat space webhook URL |
| `GOOGLE_CHAT_VERIFICATION_TOKEN` | No¹ | — | Token for verifying inbound Google Chat events |
| `PAGERDUTY_API_KEY` | No¹ | — | PagerDuty REST API key (v2) |
| `PAGERDUTY_SERVICE_ID` | No¹ | — | PagerDuty service ID used when creating incidents |
| `PAGERDUTY_FROM_EMAIL` | No¹ | — | Email address for PagerDuty `From` header (required by some accounts) |
| `OPSGENIE_API_KEY` | No¹ | — | Opsgenie API key |
| `DATADOG_API_KEY` | No¹ | — | Datadog API key |
| `DATADOG_APP_KEY` | No¹ | — | Datadog application key (required for monitor management) |
| `GRAFANA_URL` | No¹ | — | Grafana instance base URL (e.g. `https://grafana.acme.com`) |
| `GRAFANA_API_KEY` | No¹ | — | Grafana service account token or API key |
| `SNYK_API_TOKEN` | No¹ | — | Snyk personal or service account API token |
| `SNYK_ORG_ID` | No¹ | — | Snyk organisation ID |
| `SNYK_WEBHOOK_SECRET` | No¹ | — | Secret used to validate Snyk webhook HMAC-SHA256 signatures |
| `SONARQUBE_URL` | No¹ | — | SonarQube instance base URL (e.g. `https://sonar.acme.com`) |
| `SONARQUBE_TOKEN` | No¹ | — | SonarQube user or service account token |
| `SONARQUBE_WEBHOOK_SECRET` | No¹ | — | Secret used to validate SonarQube webhook HMAC-SHA256 signatures |
| `TFC_TOKEN` | No¹ | — | Terraform Cloud API token |
| `TFC_ORGANIZATION` | No¹ | — | Terraform Cloud organisation name |
| `TFC_WEBHOOK_HMAC_KEY` | No¹ | — | HMAC key for validating Terraform Cloud notification signatures |
| `ATLANTIS_URL` | No¹ | — | Atlantis server base URL (e.g. `https://atlantis.acme.com`) |
| `ATLANTIS_TOKEN` | No¹ | — | Atlantis API token (`X-Atlantis-Token`) |
| `ATLANTIS_WEBHOOK_SECRET` | No¹ | — | Secret used to validate Atlantis webhook HMAC-SHA256 signatures |

> ¹ Required only when the corresponding adapter is enabled for your deployment.
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
  # gitlab:                        # Use instead of (or alongside) github
  #   project_id: "12345"
  #   base_branch: main
  confluence:
    space_key: ACME
    agent_label: forge             # Pages with this label trigger task.created
  linear:
    team_id: "abc123"
    agent_label: forge             # Issues with this label trigger task.created
  slack:
    notification_channel: "#eng-forge"
    escalation_channel: "#eng-leads"
  # teams:
  #   notification_channel: forge-notifications
  # google_chat:
  #   space_name: spaces/XXXXXXXX

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

### Integration Adapters

Each adapter runs as an independent service and communicates with the rest of the system via the message bus. All adapters expose a `/health` endpoint and a `/webhook` endpoint for inbound events from the external platform.

| Adapter | Port | Inbound (webhooks) | Outbound (API) |
|---|---|---|---|
| **Jira** | `:8090` | `jira:issue_created` → `task.created`; `jira:issue_updated` | `GET/POST /api/v1/tickets`, `POST /api/v1/transitions` |
| **GitHub** | `:8091` | `PullRequestEvent` → `review.requested`/`task.completed`; `CheckSuiteEvent` | `POST /api/v1/branches`, `POST /api/v1/pulls`, `/api/v1/commits` |
| **Slack** | `:8092` | Slash commands, interactive payloads | Notifies on `task.completed`, `review.requested`, `escalation.created` |
| **Teams** | `:8093` | Bot Framework activity endpoint | Notifies on `task.completed`, `review.requested`, `escalation.created` |
| **Google Chat** | `:8094` | Chat event endpoint | Notifies on `task.completed`, `review.requested`, `escalation.created` |
| **GitLab** | `:8095` | `MergeEvent` → `review.requested`/`task.completed`; `PipelineEvent` → `deployment.approved`/`task.failed` | `POST /api/v1/branches`, `POST /api/v1/mergerequests`, `/api/v1/commits` |
| **Confluence** | `:8096` | `page_created` → `task.created` (when page carries `forge` label) | `GET/POST/PUT /api/v1/pages`, `GET /api/v1/spaces` |
| **Linear** | `:8097` | `Issue create` → `task.created`; `Issue update` → `task.completed`; `Issue remove` → `task.failed` | `GET/POST /api/v1/issues`, `POST /api/v1/transitions` |
| **PagerDuty** | `:8098` | `incident.trigger` → `escalation.created`; `incident.resolve` → `task.completed` | `POST /api/v1/incidents`, `PUT /api/v1/incidents?id=` |
| **Opsgenie** | `:8099` | `Create` → `escalation.created`; `Close` → `task.completed` | `POST /api/v1/alerts`, `DELETE /api/v1/alerts?id=`, `PATCH /api/v1/alerts?id=` |
| **Datadog** | `:8100` | `Triggered`/`No Data` → `escalation.created`; `Recovered` → `task.completed` | `POST /api/v1/events` |
| **Grafana** | `:8101` | `firing` → `escalation.created`; `resolved` → `task.completed` | `POST /api/v1/annotations`, `POST /api/v1/silences` |
| **Snyk** | `:8102` | `newIssues` (critical/high) → `escalation.created` | `GET /api/v1/vulnerabilities`, `GET /api/v1/projects` |
| **SonarQube** | `:8103` | `analysis.completed` quality gate `ERROR` → `escalation.created`; `FAILED`/`CANCELLED` → `task.failed` | `GET /api/v1/issues`, `GET /api/v1/qualitygates` |
| **Terraform Cloud** | `:8104` | `applied` → `task.completed`; `planned_and_finished` → `deployment.approved`; `errored`/`canceled` → `task.failed` | `GET/POST /api/v1/runs`, `GET /api/v1/workspaces` |
| **Atlantis** | `:8105` | `plan success` → `deployment.requested`; `apply success` → `task.completed`; `failure`/`error` → `task.failed` | `POST /api/v1/plan`, `POST /api/v1/apply` |

Webhook security:
- **GitHub** — HMAC-SHA256 (`GITHUB_WEBHOOK_SECRET`)
- **GitLab** — Static token header (`GITLAB_WEBHOOK_SECRET` → `X-Gitlab-Token`)
- **Linear** — HMAC-SHA256 (`LINEAR_WEBHOOK_SECRET` → `Linear-Signature`)
- **Teams** — HMAC-SHA256 (`TEAMS_HMAC_SECRET`)
- **Google Chat** — Verification token (`GOOGLE_CHAT_VERIFICATION_TOKEN`)
- **PagerDuty** — Webhook signature validation via PagerDuty's `X-PagerDuty-Signature` header (configure in PagerDuty webhook settings)
- **Opsgenie** — Webhook validation via Opsgenie's HMAC-SHA256 signature on the request body
- **Datadog** — Webhook validation via Datadog's `X-Datadog-Signature` header (configure shared secret in Datadog webhook integration settings)
- **Grafana** — Webhook validation via Grafana's `X-Grafana-Signature` header (configure shared secret in Grafana contact point settings)
- **Snyk** — HMAC-SHA256 (`SNYK_WEBHOOK_SECRET` → `X-Snyk-Signature`)
- **SonarQube** — HMAC-SHA256 (`SONARQUBE_WEBHOOK_SECRET` → `X-Sonar-Webhook-HMAC-SHA256`)
- **Terraform Cloud** — HMAC-SHA512 (`TFC_WEBHOOK_HMAC_KEY` → `X-TFE-Notification-Signature`)
- **Atlantis** — HMAC-SHA256 (`ATLANTIS_WEBHOOK_SECRET` → `X-Atlantis-Signature`)

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
