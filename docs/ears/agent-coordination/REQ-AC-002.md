# REQ-AC-002: Team Creation and Management

**Pattern:** Event-Driven
**Capability:** Agent Coordination

## Requirement

When a team is created, the system shall establish a named group of agents with shared context and communication channels.

## Acceptance Criteria

- [ ] Teams are created with a name and description
- [ ] Teams can be deleted when no longer needed
- [ ] Team members share a communication channel

## Source Evidence

- `OpenHarness/src/openharness/tools/team_create_tool.py`
- `OpenHarness/src/openharness/tools/team_delete_tool.py`
