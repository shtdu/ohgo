# REQ-AC-004: Agent Isolation

**Pattern:** Complex
**Capability:** Agent Coordination

## Requirement

If a subagent is spawned, then the system shall execute it in an isolated process with separate permissions and context, while allowing configured data sharing through team communication.

## Acceptance Criteria

- [ ] Subagents run in separate processes (subprocess backend)
- [ ] Each subagent has its own permission context
- [ ] Team-based agents share a mailbox for coordination
- [ ] Subagent failures do not crash the parent agent

## Source Evidence

- `OpenHarness/src/openharness/swarm/` — subprocess and in-process backends
