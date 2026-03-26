# Project Bootstrap Skill

## Purpose

Populate seeded `.forge/` plan document stubs with product-specific content so that all Forge agents have the context they need to begin work:
- Gather product requirements from input brief or Confluence PRD
- Generate PRODUCT.md content (vision, users, value proposition, features, success metrics)
- Generate CONTRIBUTING.md code style rules specific to the chosen tech stack
- Coordinate with the Architect agent for technical document content
- Commit populated documents via GitHub and open a PR for human review

## Prerequisites

1. Repository has been seeded with `.forge/` directory (via `forge-seeder`)
2. Product brief or Confluence PRD link provided as input
3. `GITHUB_TOKEN` with write access to the target repository
4. Architect agent available for technical document delegation

## Execution Steps

1. **Load Seeded Stubs**
   - Read existing `.forge/` documents via `plan-reader`
   - Parse `config.yaml` to extract project name, tech stack, and integration settings
   - Identify which documents still contain only placeholder comments

2. **Gather Product Context**
   - If `confluence_prd_url` is provided, fetch the PRD from Confluence
   - If `product_brief` is provided inline, use it directly
   - Extract: product purpose, target users, key features, success metrics, constraints

3. **Generate Product Documents**
   - Populate PRODUCT.md with structured product vision derived from the brief
   - Populate CONTRIBUTING.md code style section based on the tech stack in `config.yaml`
   - Ensure all `<!-- placeholder -->` comments are replaced with substantive content

4. **Delegate Technical Documents to Architect**
   - Create tasks for the Architect agent via the orchestrator:
     - `requirements-analysis` — to produce detailed functional and non-functional requirements
     - `architecture-design` — to populate ARCHITECTURE.md and DATA_MODEL.md
     - `api-design` — to populate API_CONTRACTS.md with initial endpoint specifications
   - Pass product context and `config.yaml` tech stack as input to each task
   - Set task dependencies: `requirements-analysis` must complete before `architecture-design` and `api-design`

5. **Assemble and Commit**
   - Wait for Architect agent tasks to complete
   - Collect all populated documents (PM-authored + Architect-authored)
   - Commit all documents to a `forge/pm/bootstrap-plan-docs` branch via `github-interaction`
   - Open a PR for human review with a summary of what was generated

6. **Communicate**
   - Post a summary to the project Slack channel via `slack-communication`
   - Include a link to the PR and a checklist of documents populated

## Dependencies

- `common/plan-reader` — read seeded stubs and config
- `common/github-interaction` — commit files and open PR
- `common/slack-communication` — notify team
- `common/jira-interaction` — create tracking ticket for bootstrap

## Inputs

```json
{
  "product_brief": "(required) Free-text product description, goals, and constraints",
  "confluence_prd_url": "(optional) URL to a Confluence PRD page for richer context",
  "repo": "(required) GitHub repository in org/repo format",
  "branch": "(optional) Base branch, defaults to main"
}
```

## Outputs

```json
{
  "pr_url": "https://github.com/org/repo/pull/1",
  "documents_populated": ["PRODUCT.md", "CONTRIBUTING.md", "ARCHITECTURE.md", "DATA_MODEL.md", "API_CONTRACTS.md"],
  "architect_tasks": ["task-id-1", "task-id-2", "task-id-3"],
  "tracking_ticket": "PROJ-100"
}
```

## Plan References

| Document | Sections to Consult |
|----------|---------------------|
| config.yaml | Project name, tech stack, integrations, agent list |
| PRODUCT.md | Stub to populate with product vision |
| CONTRIBUTING.md | Stub to populate with code style rules |

## Quality Gates

- [ ] All placeholder comments (`<!-- ... -->`) replaced with substantive content
- [ ] PRODUCT.md contains at minimum: overview, target users, key features, success metrics
- [ ] CONTRIBUTING.md code style section references the actual tech stack from config.yaml
- [ ] Architect agent tasks completed successfully before PR is opened
- [ ] PR opened with human reviewer assigned
- [ ] No secrets or credentials included in generated content

## Escalation Triggers

- Product brief is too vague to derive meaningful product vision (< 50 words, no clear purpose)
- Confluence PRD URL is inaccessible or returns empty content
- Architect agent tasks fail or time out
- Tech stack in config.yaml is empty or uses only defaults without customization

## Anti-Patterns

- Do not auto-merge the bootstrap PR — always require human review
- Do not generate speculative architecture without Architect agent input
- Do not populate technical documents (ARCHITECTURE.md, DATA_MODEL.md, API_CONTRACTS.md) directly — delegate to Architect
- Do not leave any placeholder comments in the final output
- Do not skip Slack notification — the team must know documents are ready for review
