# REQ-MC-002: Memory Discovery on Session Start

**Pattern:** Event-Driven
**Capability:** Memory and Context

## Requirement

When a session starts, the system shall load relevant memory files from the project and user directories into the agent's context.

## Acceptance Criteria

- [ ] Scans project-level memory directories
- [ ] Scans user-level global memory
- [ ] Loads a memory index file (MEMORY.md) summarizing available memories
- [ ] Memory content is available in the system prompt
- [ ] When memory files exist but are unreadable due to permission errors or corruption, the system logs the specific error and skips the affected entries
- [ ] When no memory files are found, the system proceeds with empty context without error

## Source Evidence

- `OpenHarness/src/openharness/memory/` — `scan_memory_files()`, `load_memory_prompt()`
