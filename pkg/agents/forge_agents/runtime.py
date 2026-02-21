"""
Forge Agent Runtime - Base framework for all agents.
"""
import os
import json
import logging
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import Any, Optional
from enum import Enum

import httpx
from anthropic import Anthropic
from pydantic import BaseModel
import redis
import yaml

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


class TaskStatus(str, Enum):
    CREATED = "created"
    QUEUED = "queued"
    EXECUTING = "executing"
    BLOCKED = "blocked"
    IN_REVIEW = "in_review"
    COMPLETED = "completed"
    FAILED = "failed"


@dataclass
class AgentIdentity:
    company_id: str
    project_id: str
    role: str
    instance_id: str

    @property
    def full_id(self) -> str:
        return f"forge:{self.company_id}:{self.project_id}:{self.role}:{self.instance_id}"


@dataclass
class Task:
    id: str
    jira_ticket_id: str
    skill_name: str
    input_payload: dict
    status: TaskStatus = TaskStatus.QUEUED
    output_payload: Optional[dict] = None
    error_message: Optional[str] = None


@dataclass
class SkillContext:
    task: Task
    plan_documents: dict[str, str]
    memories: list[dict]
    llm_config: dict


class LLMProvider:
    """Abstraction over LLM providers."""

    def __init__(self, provider: str = "anthropic", model: str = "claude-sonnet-4-5-20250929"):
        self.provider = provider
        self.model = model
        if provider == "anthropic":
            self.client = Anthropic(api_key=os.environ["ANTHROPIC_API_KEY"])

    def complete(
        self,
        system_prompt: str,
        messages: list[dict],
        max_tokens: int = 4096,
        temperature: float = 0.7,
    ) -> str:
        if self.provider == "anthropic":
            response = self.client.messages.create(
                model=self.model,
                max_tokens=max_tokens,
                system=system_prompt,
                messages=messages,
                temperature=temperature,
            )
            return response.content[0].text
        raise ValueError(f"Unknown provider: {self.provider}")


class PlanReader:
    """Reads plan documents from project repositories."""

    def __init__(self, github_adapter_url: str):
        self.github_adapter = github_adapter_url
        self.client = httpx.Client()

    def load_plans(self, repo: str, branch: str = "main") -> dict[str, str]:
        """Load all .forge/ plan documents from a repository."""
        plans = {}
        plan_files = [
            "PRODUCT.md", "ARCHITECTURE.md", "API_CONTRACTS.md",
            "DATA_MODEL.md", "UX_GUIDELINES.md", "SECURITY_POLICY.md",
            "INFRASTRUCTURE.md", "TEST_STRATEGY.md", "CONTRIBUTING.md",
            "GLOSSARY.md", "config.yaml"
        ]

        for filename in plan_files:
            path = f".forge/{filename}"
            try:
                response = self.client.get(
                    f"{self.github_adapter}/api/v1/files",
                    params={"repo": repo, "path": path, "ref": branch}
                )
                if response.status_code == 200:
                    plans[filename] = response.json()["content"]
            except Exception as e:
                logger.warning(f"Could not load {path}: {e}")

        return plans

    def get_config(self, plans: dict[str, str]) -> dict:
        """Parse the config.yaml from loaded plans."""
        if "config.yaml" in plans:
            return yaml.safe_load(plans["config.yaml"])
        return {}


class MemoryStore:
    """Medium-term memory persistence for agents."""

    def __init__(self, database_url: str):
        from sqlalchemy import create_engine
        self.engine = create_engine(database_url)

    def get_relevant_memories(
        self,
        company_id: str,
        project_id: str,
        agent_role: str,
        skill_name: str,
        limit: int = 10
    ) -> list[dict]:
        """Retrieve memories relevant to the current task."""
        from sqlalchemy import text
        with self.engine.connect() as conn:
            result = conn.execute(
                text("""
                    SELECT id, category, content, confidence, source_tickets
                    FROM agent_memory
                    WHERE company_id = :company_id
                      AND project_id = :project_id
                      AND agent_role = :agent_role
                    ORDER BY confidence DESC, last_referenced DESC
                    LIMIT :limit
                """),
                {
                    "company_id": company_id,
                    "project_id": project_id,
                    "agent_role": agent_role,
                    "limit": limit
                }
            )
            return [dict(row._mapping) for row in result]

    def store_memory(
        self,
        company_id: str,
        project_id: str,
        agent_role: str,
        category: str,
        content: str,
        source_tickets: list[str],
        confidence: float = 0.5
    ) -> str:
        """Store a new memory from task execution."""
        import uuid
        from datetime import datetime
        from sqlalchemy import text

        memory_id = str(uuid.uuid4())
        quarter = f"{datetime.now().year}-Q{(datetime.now().month - 1) // 3 + 1}"

        with self.engine.connect() as conn:
            conn.execute(
                text("""
                    INSERT INTO agent_memory
                    (id, agent_role, project_id, company_id, category, content,
                     source_tickets, confidence, quarter)
                    VALUES (:id, :agent_role, :project_id, :company_id, :category,
                            :content, :source_tickets, :confidence, :quarter)
                """),
                {
                    "id": memory_id,
                    "agent_role": agent_role,
                    "project_id": project_id,
                    "company_id": company_id,
                    "category": category,
                    "content": content,
                    "source_tickets": source_tickets,
                    "confidence": confidence,
                    "quarter": quarter
                }
            )
            conn.commit()

        return memory_id


class BaseAgent(ABC):
    """Base class for all Forge agents."""

    def __init__(
        self,
        identity: AgentIdentity,
        llm_provider: LLMProvider,
        plan_reader: PlanReader,
        memory_store: MemoryStore,
        event_bus_url: str,
    ):
        self.identity = identity
        self.llm = llm_provider
        self.plan_reader = plan_reader
        self.memory = memory_store
        self.event_bus_url = event_bus_url
        self.client = httpx.Client()
        self.skills: dict[str, "Skill"] = {}

    @property
    @abstractmethod
    def role(self) -> str:
        """Return the agent role identifier."""
        pass

    def register_skill(self, skill: "Skill"):
        """Register a skill with this agent."""
        self.skills[skill.name] = skill

    def execute_task(self, task: Task, repo: str) -> Task:
        """Execute a task using the appropriate skill."""
        logger.info(f"Agent {self.identity.full_id} executing task {task.id}")

        plans = self.plan_reader.load_plans(repo)
        memories = self.memory.get_relevant_memories(
            self.identity.company_id,
            self.identity.project_id,
            self.role,
            task.skill_name
        )

        config = self.plan_reader.get_config(plans)
        llm_config = self._get_llm_config(task.skill_name, config)

        context = SkillContext(
            task=task,
            plan_documents=plans,
            memories=memories,
            llm_config=llm_config
        )

        if task.skill_name not in self.skills:
            task.status = TaskStatus.FAILED
            task.error_message = f"Unknown skill: {task.skill_name}"
            return task

        skill = self.skills[task.skill_name]

        try:
            task.status = TaskStatus.EXECUTING
            self._publish_event("task.started", {"task_id": task.id})

            result = skill.execute(context, self.llm)

            task.output_payload = result
            task.status = TaskStatus.IN_REVIEW
            self._publish_event("task.completed", {
                "task_id": task.id,
                "output": result
            })

        except Exception as e:
            logger.exception(f"Task {task.id} failed")
            task.status = TaskStatus.FAILED
            task.error_message = str(e)
            self._publish_event("task.failed", {
                "task_id": task.id,
                "error": str(e)
            })

        return task

    def _get_llm_config(self, skill_name: str, config: dict) -> dict:
        """Determine LLM configuration for a skill."""
        return {
            "provider": "anthropic",
            "model": "claude-sonnet-4-5-20250929",
            "max_tokens": 4096,
            "temperature": 0.7
        }

    def _publish_event(self, event_type: str, payload: dict):
        """Publish an event to the message bus."""
        try:
            self.client.post(
                f"{self.event_bus_url}/publish",
                json={
                    "type": event_type,
                    "agent_id": self.identity.full_id,
                    "payload": payload
                }
            )
        except Exception as e:
            logger.error(f"Failed to publish event: {e}")


class Skill(ABC):
    """Base class for agent skills."""

    @property
    @abstractmethod
    def name(self) -> str:
        """Return the skill name identifier."""
        pass

    @property
    @abstractmethod
    def description(self) -> str:
        """Return a description of what this skill does."""
        pass

    @abstractmethod
    def execute(self, context: SkillContext, llm: LLMProvider) -> dict:
        """Execute the skill and return results."""
        pass

    def build_system_prompt(self, context: SkillContext) -> str:
        """Build the system prompt for LLM execution."""
        prompt_parts = [
            f"You are a {self.name} skill executing within the Forge agent framework.",
            f"\n## Task\n{context.task.input_payload.get('description', '')}",
        ]

        if "ARCHITECTURE.md" in context.plan_documents:
            prompt_parts.append(
                f"\n## Architecture Context\n{context.plan_documents['ARCHITECTURE.md'][:2000]}"
            )

        if context.memories:
            memory_text = "\n".join([m["content"] for m in context.memories[:5]])
            prompt_parts.append(f"\n## Learned Patterns\n{memory_text}")

        return "\n".join(prompt_parts)
