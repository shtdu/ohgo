# REQ-MC-006: Memory Size Limits

**Pattern:** Event-Driven
**Capability:** Memory and Context

## Requirement

When a memory entry is added, the system shall enforce configurable limits on the number and size of memory entries to prevent unbounded growth.

## Acceptance Criteria

- [ ] Maximum number of memory files is configurable (default: 200 files)
- [ ] Maximum content size per entry is configurable (default: 32KB per entry)
- [ ] When a memory write would exceed a configured limit, the system rejects the write and reports the limit condition to the agent

## Source Evidence

- `OpenHarness/src/openharness/settings.py` — `memory.max_files`, `memory.max_entrypoint_lines`
