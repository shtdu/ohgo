# REQ-AT-005: Task Progress Tracking

**Pattern:** Event-Driven
**Capability:** Task Automation

## Requirement

When a task updates its progress, the system shall persist the progress metadata (percentage, status note) for status queries.

## Acceptance Criteria

- [ ] Progress is a percentage value (0-100)
- [ ] A status note describes current activity
- [ ] Progress metadata is persisted and available to task retrieval operations (REQ-AT-002, REQ-AT-004)
- [ ] When an invalid progress value is provided (negative, over 100, or non-numeric), the system rejects the update and returns a validation error

## Source Evidence

- `OpenHarness/src/openharness/tools/task_update_tool.py` — `progress`, `status_note` parameters
