# REQ-MC-006: Memory Size Limits

**Pattern:** State-Driven
**Capability:** Memory and Context

## Requirement

While the memory system is active, the system shall enforce configurable limits on the number and size of memory files to prevent unbounded growth.

## Acceptance Criteria

- [ ] Maximum number of memory files is configurable
- [ ] Maximum entry point lines is configurable
- [ ] The system warns when approaching limits

## Source Evidence

- `OpenHarness/src/openharness/settings.py` — `memory.max_files`, `memory.max_entrypoint_lines`
