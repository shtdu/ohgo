# REQ-AT-005: Task Progress Tracking

**Pattern:** Event-Driven
**Capability:** Task Automation

## Requirement

When a task updates its progress, the system shall persist the progress metadata (percentage, status note) for status queries.

## Acceptance Criteria

- [ ] Progress is a percentage value (0-100)
- [ ] A status note describes current activity
- [ ] Progress is queryable via task list and task get

## Source Evidence

- `OpenHarness/src/openharness/tools/task_update_tool.py` — `progress`, `status_note` parameters
