# REQ-AT-001: Background Task Execution

**Pattern:** Event-Driven
**Capability:** Task Automation

## Requirement

When a background task is created, the system shall execute it independently of the main conversation.

## Acceptance Criteria

- [ ] Tasks can run external commands
- [ ] Tasks can run agent-driven prompts
- [ ] Tasks execute independently of the main conversation loop
- [ ] Task state is queryable at any time
- [ ] When task execution fails (command not found, timeout), the task state transitions to failed with an error message

## Source Evidence

- `OpenHarness/src/openharness/tools/task_create_tool.py`
- `OpenHarness/src/openharness/tasks/`
