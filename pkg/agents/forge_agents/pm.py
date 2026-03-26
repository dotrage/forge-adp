"""
PM Agent implementation.

Four skills:
  - sprint-planning      : decompose goals into sprint-ready Jira tickets
  - status-reporting     : generate sprint status reports
  - ticket-triage        : review and route incoming tickets
  - project-bootstrap    : populate seeded .forge/ plan documents for a new project
"""
import json
import re
from datetime import datetime, timezone

from .runtime import BaseAgent, Skill, SkillContext, LLMProvider


# ---------------------------------------------------------------------------
# project-bootstrap
# ---------------------------------------------------------------------------

class ProjectBootstrapSkill(Skill):
    """Populate seeded .forge/ plan documents with product-specific content."""

    @property
    def name(self) -> str:
        return "project-bootstrap"

    @property
    def description(self) -> str:
        return "Populate seeded plan document stubs with product vision and contributing guidelines"

    def execute(self, context: SkillContext, llm: LLMProvider) -> dict:
        product_brief = context.task.input_payload.get("product_brief", "")
        repo = context.task.input_payload.get("repo", "")
        config_yaml = context.plan_documents.get("config.yaml", "")
        product_stub = context.plan_documents.get("PRODUCT.md", "")
        contributing_stub = context.plan_documents.get("CONTRIBUTING.md", "")

        if not product_brief:
            return {
                "error": "product_brief is required",
                "escalation": "Product brief is too vague or missing"
            }

        system_prompt = self.build_system_prompt(context)
        system_prompt += (
            "\n\nYou are the Forge PM Agent running the project-bootstrap skill. "
            "Your job is to populate PRODUCT.md and CONTRIBUTING.md with substantive "
            "content derived from the product brief. Be concrete and specific. "
            "Replace all placeholder comments with real content."
        )

        messages = [
            {
                "role": "user",
                "content": f"""
Populate the plan documents for this new project.

## Product Brief
{product_brief}

## Project Config
{config_yaml[:2000] if config_yaml else "(not available)"}

## Current PRODUCT.md Stub
{product_stub}

## Current CONTRIBUTING.md Stub
{contributing_stub}

## Instructions

### PRODUCT.md
Populate every section with substantive content:
1. **Overview** — What the product is and why it exists
2. **Target Users** — Who uses it and their key characteristics
3. **Core Value Proposition** — The problem it solves and why users choose it
4. **Key Features** — Bulleted list of major capabilities
5. **Success Metrics** — Measurable KPIs

### CONTRIBUTING.md
Keep the existing testing requirements, PR template, branch naming, and commit message sections.
Populate the **Code Style** section with specific rules for the tech stack in config.yaml.

Replace ALL `<!-- ... -->` placeholder comments with real content.

Return strictly valid JSON:
{{
  "product_md": "Full markdown content for PRODUCT.md",
  "contributing_md": "Full markdown content for CONTRIBUTING.md",
  "architect_tasks": [
    {{
      "skill_name": "requirements-analysis",
      "input": {{
        "product_brief": "...",
        "repo": "..."
      }}
    }},
    {{
      "skill_name": "architecture-design",
      "input": {{
        "product_brief": "...",
        "repo": "..."
      }},
      "depends_on": "requirements-analysis"
    }},
    {{
      "skill_name": "api-design",
      "input": {{
        "product_brief": "...",
        "repo": "..."
      }},
      "depends_on": "requirements-analysis"
    }}
  ]
}}
""",
            }
        ]

        response = llm.complete(system_prompt, messages, max_tokens=8000, temperature=0.4)
        result = self._parse_json_response(response)

        # Attach metadata for the orchestrator to create architect tasks
        result["repo"] = repo
        result["documents_populated"] = ["PRODUCT.md", "CONTRIBUTING.md"]
        result.setdefault("generated_at", datetime.now(timezone.utc).isoformat())
        return result

    def _parse_json_response(self, text: str) -> dict:
        match = re.search(r"```(?:json)?\s*(\{.*?\})\s*```", text, re.DOTALL)
        if match:
            text = match.group(1)
        try:
            return json.loads(text)
        except json.JSONDecodeError:
            return {"raw_response": text, "parse_error": True}


# ---------------------------------------------------------------------------
# sprint-planning
# ---------------------------------------------------------------------------

class SprintPlanningSkill(Skill):
    """Transform product goals into sprint-ready Jira tickets."""

    @property
    def name(self) -> str:
        return "sprint-planning"

    @property
    def description(self) -> str:
        return "Decompose product goals into structured, sprint-ready Jira tickets"

    def execute(self, context: SkillContext, llm: LLMProvider) -> dict:
        goal = context.task.input_payload.get("goal_description", "")
        capacity = context.task.input_payload.get("sprint_capacity", None)
        arch_doc = context.plan_documents.get("ARCHITECTURE.md", "")
        contributing_doc = context.plan_documents.get("CONTRIBUTING.md", "")

        system_prompt = self.build_system_prompt(context)
        system_prompt += (
            "\n\nYou are the Forge PM Agent running the sprint-planning skill. "
            "Decompose the goal into specific, estimable Jira tickets. Every ticket "
            "must have acceptance criteria. Respect sprint capacity limits."
        )

        messages = [
            {
                "role": "user",
                "content": f"""
Decompose the following goal into sprint-ready tickets.

## Goal
{goal}

## Sprint Capacity
{f"{capacity} story points" if capacity else "(not specified)"}

## Architecture Context
{arch_doc[:2000] if arch_doc else "(not available)"}

## Contributing Guidelines
{contributing_doc[:1000] if contributing_doc else "(not available)"}

## Instructions
1. Identify Epics or Stories from the goal.
2. Break each Story into Tasks across agent roles (backend, frontend, QA, etc.).
3. Assign story point estimates and priorities.
4. Set dependency links between tickets.

Return strictly valid JSON:
{{
  "tickets": [
    {{
      "type": "Story|Task",
      "summary": "...",
      "description": "...",
      "acceptance_criteria": ["..."],
      "estimate_points": 0,
      "priority": "High|Medium|Low",
      "agent_role": "...",
      "depends_on": []
    }}
  ],
  "total_estimate_points": 0,
  "sprint_plan_summary": "..."
}}
""",
            }
        ]

        response = llm.complete(system_prompt, messages, max_tokens=8000, temperature=0.4)
        result = self._parse_json_response(response)
        result.setdefault("generated_at", datetime.now(timezone.utc).isoformat())
        return result

    def _parse_json_response(self, text: str) -> dict:
        match = re.search(r"```(?:json)?\s*(\{.*?\})\s*```", text, re.DOTALL)
        if match:
            text = match.group(1)
        try:
            return json.loads(text)
        except json.JSONDecodeError:
            return {"raw_response": text, "parse_error": True}


# ---------------------------------------------------------------------------
# status-reporting
# ---------------------------------------------------------------------------

class StatusReportingSkill(Skill):
    """Generate structured sprint status reports."""

    @property
    def name(self) -> str:
        return "status-reporting"

    @property
    def description(self) -> str:
        return "Generate sprint status reports from Jira data"

    def execute(self, context: SkillContext, llm: LLMProvider) -> dict:
        sprint_id = context.task.input_payload.get("sprint_id", "")
        audience = context.task.input_payload.get("report_audience", "team")

        system_prompt = self.build_system_prompt(context)
        system_prompt += (
            "\n\nYou are the Forge PM Agent running the status-reporting skill. "
            f"Generate a status report for audience: {audience}. "
            "Be data-driven. Highlight blockers and risks prominently."
        )

        messages = [
            {
                "role": "user",
                "content": f"""
Generate a sprint status report for sprint `{sprint_id}`.

## Audience
{audience} (adjust detail level accordingly)

## Instructions
1. Categorise tickets: To Do, In Progress, In Review, Done, Blocked.
2. Calculate completion percentage and velocity.
3. Identify risks and blockers.
4. Produce a narrative summary appropriate for the audience.

Return strictly valid JSON:
{{
  "sprint_id": "{sprint_id}",
  "completion_pct": 0.0,
  "velocity": 0,
  "categories": {{"todo": 0, "in_progress": 0, "in_review": 0, "done": 0, "blocked": 0}},
  "blockers": ["..."],
  "risks": ["..."],
  "narrative": "...",
  "report_markdown": "..."
}}
""",
            }
        ]

        response = llm.complete(system_prompt, messages, max_tokens=6000, temperature=0.3)
        result = self._parse_json_response(response)
        result.setdefault("generated_at", datetime.now(timezone.utc).isoformat())
        return result

    def _parse_json_response(self, text: str) -> dict:
        match = re.search(r"```(?:json)?\s*(\{.*?\})\s*```", text, re.DOTALL)
        if match:
            text = match.group(1)
        try:
            return json.loads(text)
        except json.JSONDecodeError:
            return {"raw_response": text, "parse_error": True}


# ---------------------------------------------------------------------------
# ticket-triage
# ---------------------------------------------------------------------------

class TicketTriageSkill(Skill):
    """Review incoming tickets for completeness and route to agents."""

    @property
    def name(self) -> str:
        return "ticket-triage"

    @property
    def description(self) -> str:
        return "Triage incoming Jira tickets — verify completeness, assign priority, route to agents"

    def execute(self, context: SkillContext, llm: LLMProvider) -> dict:
        ticket_ids = context.task.input_payload.get("ticket_ids", [])

        system_prompt = self.build_system_prompt(context)
        system_prompt += (
            "\n\nYou are the Forge PM Agent running the ticket-triage skill. "
            "Review each ticket for completeness (summary, description, acceptance criteria). "
            "Assign priority based on business impact. Route to the correct agent role."
        )

        messages = [
            {
                "role": "user",
                "content": f"""
Triage the following tickets: {json.dumps(ticket_ids)}

## Instructions
1. For each ticket, verify completeness (summary, description, acceptance criteria, type, priority).
2. Assign priority based on business impact and dependencies.
3. Route to the correct agent role based on ticket type and labels.
4. Flag incomplete tickets with specific clarification questions.

Return strictly valid JSON:
{{
  "triage_results": [
    {{
      "ticket_id": "...",
      "is_complete": true,
      "assigned_priority": "Critical|High|Medium|Low",
      "assigned_agent_role": "...",
      "missing_fields": [],
      "clarification_questions": []
    }}
  ],
  "incomplete_count": 0,
  "routed_count": 0
}}
""",
            }
        ]

        response = llm.complete(system_prompt, messages, max_tokens=6000, temperature=0.3)
        result = self._parse_json_response(response)
        result.setdefault("generated_at", datetime.now(timezone.utc).isoformat())
        return result

    def _parse_json_response(self, text: str) -> dict:
        match = re.search(r"```(?:json)?\s*(\{.*?\})\s*```", text, re.DOTALL)
        if match:
            text = match.group(1)
        try:
            return json.loads(text)
        except json.JSONDecodeError:
            return {"raw_response": text, "parse_error": True}


# ---------------------------------------------------------------------------
# Agent
# ---------------------------------------------------------------------------

class PMAgent(BaseAgent):
    """
    PM Agent — product management, sprint planning, and project bootstrap.

    Decomposes product goals into actionable tickets, triages incoming work,
    reports on sprint progress, and orchestrates initial project setup by
    collaborating with the Architect agent to populate plan documents.
    """

    @property
    def role(self) -> str:
        return "pm"

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.register_skill(ProjectBootstrapSkill())
        self.register_skill(SprintPlanningSkill())
        self.register_skill(StatusReportingSkill())
        self.register_skill(TicketTriageSkill())
