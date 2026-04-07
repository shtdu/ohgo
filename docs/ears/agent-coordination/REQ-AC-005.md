# REQ-AC-005: Agent Task Lifecycle

**Pattern:** Ubiquitous
**Capability:** Agent Coordination

## Requirement

The system shall manage the full lifecycle of subagent tasks including creation, execution, output collection, and termination.

## Acceptance Criteria

- [ ] Tasks are created with a unique ID
- [ ] Task state is trackable (pending, running, completed, failed)
- [ ] Task output can be retrieved after completion
- [ ] Tasks can be stopped on user request

## Source Evidence

- `OpenHarness/src/openharness/tools/task_create_tool.py`
- `OpenHarness/src/openharness/tools/task_get_tool.py`
- `OpenHarness/src/openharness/tools/task_output_tool.py`
- `OpenHarness/src/openharness/tools/task_stop_tool.py`
- `OpenHarness/src/openharness/tools/task_update_tool.py`
