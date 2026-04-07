# REQ-AT-001: Background Task Execution

**Pattern:** Ubiquitous
**Capability:** Task Automation

## Requirement

The system shall support background task execution for both bash commands and agent prompts, allowing work to proceed independently of the main conversation.

## Acceptance Criteria

- [ ] Supports local bash tasks (shell commands)
- [ ] Supports local agent tasks (subagent prompts)
- [ ] Tasks execute independently of the main conversation loop
- [ ] Task state is queryable at any time

## Source Evidence

- `OpenHarness/src/openharness/tools/task_create_tool.py`
- `OpenHarness/src/openharness/tasks/`
