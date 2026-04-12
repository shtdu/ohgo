# REQ-AC-004: Agent Isolation

**Pattern:** State-Driven
**Capability:** Agent Coordination

## Requirement

While a subagent is running, the system shall maintain execution isolation with separate permissions and context from the parent agent.

## Acceptance Criteria

- [ ] Subagent execution is isolated from the parent agent, such that a subagent failure does not affect the parent
- [ ] Each subagent operates under its own permission rules
- [ ] A subagent cannot read or modify the parent agent's conversation history or tool state
- [ ] When a subagent attempts to access resources outside its scope (file paths, environment variables), the system blocks the access and reports the violation to the parent agent

## Source Evidence

- `OpenHarness/src/openharness/swarm/` — subprocess and in-process backends
