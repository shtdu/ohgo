# REQ-AC-001: Subagent Spawning

**Pattern:** Event-Driven
**Capability:** Agent Coordination

## Requirement

When the agent delegates work to a subagent, the system shall create a new agent instance with an isolated context and execute the specified task.

## Acceptance Criteria

- [ ] Accepts a task description and prompt for the subagent
- [ ] Supports specifying subagent type (general-purpose, specialized)
- [ ] Supports model selection per subagent
- [ ] Subagent output is included in the parent agent's conversation as a tool result
- [ ] When subagent spawning fails due to resource limits or invalid configuration, the system returns an error to the parent agent

## Source Evidence

- `OpenHarness/src/openharness/tools/agent_tool.py`
- `OpenHarness/src/openharness/swarm/` — swarm/subprocess execution
