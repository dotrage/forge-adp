"""
Architect Agent implementation.

Three skills:
  - requirements-analysis  : translate product brief into structured requirements
  - architecture-design    : design system architecture and data model
  - api-design             : design API surface and endpoint specifications
"""
import json
import re
from datetime import datetime, timezone

from .runtime import BaseAgent, Skill, SkillContext, LLMProvider


# ---------------------------------------------------------------------------
# requirements-analysis
# ---------------------------------------------------------------------------

class RequirementsAnalysisSkill(Skill):
    """Translate a product brief into structured functional and non-functional requirements."""

    @property
    def name(self) -> str:
        return "requirements-analysis"

    @property
    def description(self) -> str:
        return "Extract functional and non-functional requirements from a product brief"

    def execute(self, context: SkillContext, llm: LLMProvider) -> dict:
        product_brief = context.task.input_payload.get("product_brief", "")
        product_md = context.plan_documents.get("PRODUCT.md", "")
        config_yaml = context.plan_documents.get("config.yaml", "")

        system_prompt = self.build_system_prompt(context)
        system_prompt += (
            "\n\nYou are the Forge Architect Agent running the requirements-analysis skill. "
            "Your job is to extract structured, testable requirements from the product brief. "
            "Be thorough but do not invent features not present in the brief. "
            "Flag assumptions and open questions explicitly."
        )

        messages = [
            {
                "role": "user",
                "content": f"""
Analyse the following product brief and produce structured requirements.

## Product Brief
{product_brief}

## Product Vision (if available)
{product_md[:3000] if product_md else "(not yet populated)"}

## Project Config
{config_yaml[:2000] if config_yaml else "(not available)"}

## Instructions
1. Extract functional requirements — each must be specific, testable, and prioritised (Must/Should/Could).
2. Define non-functional requirements covering at minimum: performance, security, availability, scalability.
3. List constraints imposed by the chosen tech stack.
4. List assumptions you are making.
5. Surface open questions that need product owner clarification.

Return strictly valid JSON:
{{
  "functional_requirements": [
    {{
      "id": "REQ-001",
      "description": "...",
      "priority": "Must|Should|Could",
      "acceptance_criteria": ["..."],
      "source": "product_brief"
    }}
  ],
  "non_functional_requirements": [
    {{
      "id": "NFR-001",
      "category": "performance|security|scalability|availability|compliance",
      "description": "...",
      "target": "...",
      "rationale": "..."
    }}
  ],
  "constraints": ["..."],
  "assumptions": ["..."],
  "open_questions": ["..."]
}}
""",
            }
        ]

        response = llm.complete(system_prompt, messages, max_tokens=8000, temperature=0.3)
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
# architecture-design
# ---------------------------------------------------------------------------

class ArchitectureDesignSkill(Skill):
    """Design system architecture and data model, populating ARCHITECTURE.md and DATA_MODEL.md."""

    @property
    def name(self) -> str:
        return "architecture-design"

    @property
    def description(self) -> str:
        return "Design system architecture and data model from requirements"

    def execute(self, context: SkillContext, llm: LLMProvider) -> dict:
        requirements = context.task.input_payload.get("requirements", {})
        product_brief = context.task.input_payload.get("product_brief", "")
        config_yaml = context.plan_documents.get("config.yaml", "")
        arch_stub = context.plan_documents.get("ARCHITECTURE.md", "")
        data_model_stub = context.plan_documents.get("DATA_MODEL.md", "")

        system_prompt = self.build_system_prompt(context)
        system_prompt += (
            "\n\nYou are the Forge Architect Agent running the architecture-design skill. "
            "Design a practical, right-sized architecture. Match complexity to requirements — "
            "do not over-architect. Every decision must include rationale."
        )

        messages = [
            {
                "role": "user",
                "content": f"""
Design the system architecture and data model for this project.

## Product Brief
{product_brief[:2000]}

## Requirements
{json.dumps(requirements, indent=2)[:4000]}

## Project Config
{config_yaml[:2000] if config_yaml else "(not available)"}

## Current ARCHITECTURE.md Stub
{arch_stub[:1000]}

## Current DATA_MODEL.md Stub
{data_model_stub[:500]}

## Instructions
1. Write a complete ARCHITECTURE.md covering: system overview, tech stack (use values from config), service boundaries, API design patterns, error handling, security (auth/authz), and data flow.
2. Write a complete DATA_MODEL.md covering: overview, entity definitions with attributes and types, entity relationships, and migration strategy.
3. List key architecture decisions with rationale and alternatives considered.
4. Replace ALL placeholder comments with substantive content.

Return strictly valid JSON:
{{
  "architecture_md": "Full markdown content for ARCHITECTURE.md",
  "data_model_md": "Full markdown content for DATA_MODEL.md",
  "entities": [
    {{
      "name": "...",
      "attributes": [{{"name": "...", "type": "...", "constraints": "..."}}],
      "relationships": [{{"target": "...", "type": "one-to-many", "description": "..."}}]
    }}
  ],
  "architecture_decisions": [
    {{
      "decision": "...",
      "rationale": "...",
      "alternatives_considered": ["..."]
    }}
  ]
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
# api-design
# ---------------------------------------------------------------------------

class APIDesignSkill(Skill):
    """Design the API surface and populate API_CONTRACTS.md."""

    @property
    def name(self) -> str:
        return "api-design"

    @property
    def description(self) -> str:
        return "Design RESTful API endpoints and populate API contracts"

    def execute(self, context: SkillContext, llm: LLMProvider) -> dict:
        requirements = context.task.input_payload.get("requirements", {})
        entities = context.task.input_payload.get("entities", [])
        product_brief = context.task.input_payload.get("product_brief", "")
        config_yaml = context.plan_documents.get("config.yaml", "")
        api_stub = context.plan_documents.get("API_CONTRACTS.md", "")

        system_prompt = self.build_system_prompt(context)
        system_prompt += (
            "\n\nYou are the Forge Architect Agent running the api-design skill. "
            "Design a consistent, RESTful API surface. Preserve the existing response envelope "
            "and error format from the stub. Every endpoint must specify auth requirements."
        )

        messages = [
            {
                "role": "user",
                "content": f"""
Design the API surface for this project.

## Product Brief
{product_brief[:2000]}

## Requirements
{json.dumps(requirements, indent=2)[:4000]}

## Known Entities
{json.dumps(entities, indent=2)[:2000] if entities else "(pending from architecture-design)"}

## Project Config
{config_yaml[:2000] if config_yaml else "(not available)"}

## Current API_CONTRACTS.md Stub
{api_stub}

## Instructions
1. Preserve the existing versioning strategy, response envelope, error format, and status codes table from the stub.
2. Design RESTful endpoints for each core resource derived from the requirements.
3. For each endpoint specify: method, path, description, auth requirement, request schema, response schema, and applicable status codes.
4. Include pagination parameters for list endpoints.
5. Include at least one request/response example per resource group.
6. Replace the placeholder comment in the Endpoints section with full specifications.

Return strictly valid JSON:
{{
  "api_contracts_md": "Full markdown content for API_CONTRACTS.md",
  "endpoints": [
    {{
      "method": "GET|POST|PUT|PATCH|DELETE",
      "path": "/api/v1/...",
      "description": "...",
      "auth_required": true,
      "request_schema": {{}},
      "response_schema": {{}},
      "status_codes": [200, 400, 401, 404]
    }}
  ],
  "resource_count": 0,
  "endpoint_count": 0
}}
""",
            }
        ]

        response = llm.complete(system_prompt, messages, max_tokens=8000, temperature=0.3)
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

class ArchitectAgent(BaseAgent):
    """
    Architect Agent — technical architecture, requirements analysis, and API design.

    Collaborates with the PM agent during project bootstrap to translate product
    vision into technical plans. Produces ARCHITECTURE.md, DATA_MODEL.md, and
    API_CONTRACTS.md content. Does not implement code directly; provides the
    technical foundation that development agents build upon.
    """

    @property
    def role(self) -> str:
        return "architect"

    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.register_skill(RequirementsAnalysisSkill())
        self.register_skill(ArchitectureDesignSkill())
        self.register_skill(APIDesignSkill())
