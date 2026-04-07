# REQ-AC-005: Agent Task Lifecycle

**Pattern:** Event-Driven
**Capability:** Agent Coordination

## Requirement

When a subagent task is created, the system shall manage its lifecycle including execution in coordination with teams, output relay to the parent agent, and termination.

## Acceptance Criteria

- [ ] Subagent output is relayed to the parent agent upon completion
- [ ] Subagent execution state (running, completed, failed) is observable via the background task query tools (per REQ-AT-002)
- [ ] When a subagent task completes, its result is distinguishable from other background task results as originating from a subagent
- [ ] Subagent tasks can be stopped on user request, relayed through the background task system (per REQ-AT-002)

## Source Evidence

- `OpenHarness/src/openharness/swarm/` — subagent spawning and management
