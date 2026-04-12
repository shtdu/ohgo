# REQ-MC-005: Memory Entry Management

**Pattern:** Event-Driven
**Capability:** Memory and Context

## Requirement

When the agent adds, removes, or updates a memory entry, the system shall persist the change to the memory file system and update the memory index.

## Acceptance Criteria

- [ ] Adding memory creates a new markdown file with frontmatter
- [ ] Removing memory deletes the corresponding file
- [ ] The memory index (MEMORY.md) is updated to reflect changes
- [ ] When a memory write fails (disk full, permission denied), the system reports the error to the user and retains existing memory entries

## Source Evidence

- `OpenHarness/src/openharness/memory/` — `add_memory_entry()`, `remove_memory_entry()`
