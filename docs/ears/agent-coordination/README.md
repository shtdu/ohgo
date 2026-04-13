# Agent Coordination

How multiple agents are spawned, coordinated, and communicate to accomplish tasks cooperatively.

## Requirements

| ID | Title | Pattern |
|----|-------|---------|
| [REQ-AC-001](#req-ac-001-subagent-spawning) | Subagent Spawning | Event-Driven |
| [REQ-AC-002](#req-ac-002-team-creation-and-management) | Team Creation and Management | Event-Driven |
| [REQ-AC-003](#req-ac-003-inter-agent-messaging) | Inter-Agent Messaging | Event-Driven |
| [REQ-AC-004](#req-ac-004-agent-isolation) | Agent Isolation | State-Driven |
| [REQ-AC-005](#req-ac-005-agent-task-lifecycle) | Agent Task Lifecycle | Event-Driven |

## Dependencies

Cross-references to other domains:
- [Automation](../automation/README.md)

## Details

## REQ-AC-001: Subagent Spawning

**Pattern:** Event-Driven

### Requirement

When the agent delegates work to a subagent, the system shall create a new agent instance with an isolated context and execute the specified task.

### Acceptance Criteria

- [ ] Accepts a task description and prompt for the subagent
- [ ] Supports specifying subagent type (general-purpose, specialized)
- [ ] Supports model selection per subagent
- [ ] Subagent output is included in the parent agent's conversation as a tool result
- [ ] When subagent spawning fails due to resource limits or invalid configuration, the system returns an error to the parent agent


---

## REQ-AC-002: Team Creation and Management

**Pattern:** Event-Driven

### Requirement

When a team is created, the system shall establish a named group of agents with shared context and communication channels.

### Acceptance Criteria

- [ ] Teams are created with a name and description
- [ ] A created team persists until explicitly disbanded
- [ ] Team members share a communication channel
- [ ] Creating a team with a name that already exists produces an error


---

## REQ-AC-003: Inter-Agent Messaging

**Pattern:** Event-Driven

### Requirement

When an agent sends a message to another agent, the system shall deliver the message to the target agent's mailbox.

### Acceptance Criteria

- [ ] Messages are addressed by task ID
- [ ] The target agent receives the message in its input stream
- [ ] The sending agent continues execution without waiting for the receiving agent to process the message
- [ ] When a message is sent to a non-existent or terminated agent, the system returns a delivery failure error to the sender


---

## REQ-AC-004: Agent Isolation

**Pattern:** State-Driven

### Requirement

While a subagent is running, the system shall maintain execution isolation with separate permissions and context from the parent agent.

### Acceptance Criteria

- [ ] Subagent execution is isolated from the parent agent, such that a subagent failure does not affect the parent
- [ ] Each subagent operates under its own permission rules
- [ ] A subagent cannot read or modify the parent agent's conversation history or tool state
- [ ] When a subagent attempts to access resources outside its scope (file paths, environment variables), the system blocks the access and reports the violation to the parent agent


---

## REQ-AC-005: Agent Task Lifecycle

**Pattern:** Event-Driven

### Requirement

When a subagent task is created, the system shall manage its lifecycle including execution in coordination with teams, output relay to the parent agent, and termination.

### Acceptance Criteria

- [ ] Subagent output is relayed to the parent agent upon completion
- [ ] Subagent execution state (running, completed, failed) is observable via the background task query tools (per REQ-AT-002)
- [ ] When a subagent task completes, its result is distinguishable from other background task results as originating from a subagent
- [ ] Subagent tasks can be stopped on user request, relayed through the background task system (per REQ-AT-002)
