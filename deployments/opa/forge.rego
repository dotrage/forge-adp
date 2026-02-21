package forge.authz

default allow = false

# Protocol-level hard limits - never overridable
deny {
    input.action == "force_push"
    input.resource == "main"
}

deny {
    input.action == "delete"
    input.resource == "production_database"
}

deny {
    input.action == "disable"
    input.resource == "security_scan"
}

# Agent role permissions
allow {
    not deny
    agent_has_permission
}

agent_has_permission {
    input.action == "create_pr"
    role := get_agent_role(input.agent_id)
    role in ["backend-developer", "frontend-developer", "dba", "devops", "qa"]
}

agent_has_permission {
    input.action == "post_review"
    role := get_agent_role(input.agent_id)
    role in ["secops", "qa"]
}

agent_has_permission {
    input.action == "create_ticket"
    role := get_agent_role(input.agent_id)
    role == "pm"
}

agent_has_permission {
    input.action == "propose_migration"
    role := get_agent_role(input.agent_id)
    role == "dba"
}

# Helper to extract role from agent ID
get_agent_role(agent_id) = role {
    parts := split(agent_id, ":")
    role := parts[3]
}
