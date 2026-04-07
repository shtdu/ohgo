# REQ-MC-002: Memory Discovery on Session Start

**Pattern:** Event-Driven
**Capability:** Memory and Context

## Requirement

When a session starts, the system shall discover and load relevant memory files from the project and user directories into the agent's context.

## Acceptance Criteria

- [ ] Scans project-level memory directories
- [ ] Scans user-level global memory
- [ ] Loads a memory index file (MEMORY.md) summarizing available memories
- [ ] Memory content is available in the system prompt

## Source Evidence

- `OpenHarness/src/openharness/memory/` — `scan_memory_files()`, `load_memory_prompt()`
