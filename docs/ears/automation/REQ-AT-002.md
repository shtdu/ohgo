# REQ-AT-002: Background Task Lifecycle Management

**Pattern:** Ubiquitous
**Capability:** Task Automation

## Requirement

The system shall manage the full lifecycle of background tasks: creation, execution, output streaming, progress tracking, and termination.

## Acceptance Criteria

- [ ] Tasks are created with unique IDs
- [ ] Task state transitions: pending → running → completed/failed
- [ ] Output can be streamed or retrieved on demand
- [ ] Tasks can be stopped by user request
- [ ] Task progress can be updated with metadata

## Source Evidence

- `OpenHarness/src/openharness/tools/task_create_tool.py`
- `OpenHarness/src/openharness/tools/task_get_tool.py`
- `OpenHarness/src/openharness/tools/task_list_tool.py`
- `OpenHarness/src/openharness/tools/task_output_tool.py`
- `OpenHarness/src/openharness/tools/task_stop_tool.py`
- `OpenHarness/src/openharness/tools/task_update_tool.py`
