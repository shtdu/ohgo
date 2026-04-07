# REQ-AT-002: Background Task Lifecycle Management

**Pattern:** State-Driven
**Capability:** Task Automation

## Requirement

While a background task exists, the system shall manage its lifecycle through creation, execution, output streaming, progress tracking, and termination.

## Acceptance Criteria

- [ ] Tasks are created with unique IDs
- [ ] Task state transitions follow: pending, running, completed, or failed
- [ ] Task output is retrievable while the task exists
- [ ] Tasks can be stopped by user request
- [ ] Task progress percentage and status are queryable

## Source Evidence

- `OpenHarness/src/openharness/tools/task_create_tool.py`
- `OpenHarness/src/openharness/tools/task_get_tool.py`
- `OpenHarness/src/openharness/tools/task_list_tool.py`
- `OpenHarness/src/openharness/tools/task_output_tool.py`
- `OpenHarness/src/openharness/tools/task_stop_tool.py`
- `OpenHarness/src/openharness/tools/task_update_tool.py`
