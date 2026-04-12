# REQ-AT-002: Background Task Lifecycle Management

**Pattern:** State-Driven
**Capability:** Task Automation

## Requirement

While a background task exists, the system shall manage its complete lifecycle from creation through termination. Progress tracking is covered separately by REQ-AT-005.

## Acceptance Criteria

- [ ] Tasks are created with unique IDs
- [ ] Task state transitions follow: pending, running, completed, or failed
- [ ] Tasks can be stopped by user request
- [ ] When a task creation fails (duplicate ID, invalid state), the system returns a descriptive error identifying the cause

## Source Evidence

- `OpenHarness/src/openharness/tools/task_create_tool.py`
- `OpenHarness/src/openharness/tools/task_get_tool.py`
- `OpenHarness/src/openharness/tools/task_list_tool.py`
- `OpenHarness/src/openharness/tools/task_output_tool.py`
- `OpenHarness/src/openharness/tools/task_stop_tool.py`
- `OpenHarness/src/openharness/tools/task_update_tool.py`
