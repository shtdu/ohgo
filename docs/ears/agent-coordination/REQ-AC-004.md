# REQ-AC-004: Agent Isolation

**Pattern:** State-Driven
**Capability:** Agent Coordination

## Requirement

While a subagent is running, the system shall maintain execution isolation with separate permissions and context, while allowing communication through any team the subagent belongs to.

## Acceptance Criteria

- [ ] Subagent execution is isolated from the parent agent, such that a subagent failure does not affect the parent
- [ ] Each subagent operates under its own permission rules
- [ ] Team-based agents communicate through the messaging system (Agent Coordination domain)
- [ ] A subagent cannot read or modify the parent agent's conversation history or tool state

## Source Evidence

- `OpenHarness/src/openharness/swarm/` — subprocess and in-process backends
