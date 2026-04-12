# REQ-AT-004: Task Output Retrieval

**Pattern:** Event-Driven
**Capability:** Task Automation

## Requirement

When the user requests task output, the system shall return the accumulated output up to a configurable size limit.

## Acceptance Criteria

- [ ] Returns output from completed or running tasks
- [ ] Respects a configurable maximum byte limit
- [ ] When output exceeds the size limit, the returned content is truncated and includes a note indicating truncation and the original size
- [ ] When output is requested for an invalid or expired task ID, the system returns an error indicating the task was not found

## Source Evidence

- `OpenHarness/src/openharness/tools/task_output_tool.py` — `max_bytes` parameter
