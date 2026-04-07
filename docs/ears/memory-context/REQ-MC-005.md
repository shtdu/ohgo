# REQ-MC-005: Memory Entry Management

**Pattern:** Event-Driven
**Capability:** Memory and Context

## Requirement

When the agent adds or removes a memory entry, the system shall persist the change to the memory file system and update the memory index.

## Acceptance Criteria

- [ ] Adding memory creates a new markdown file with frontmatter
- [ ] Removing memory deletes the corresponding file
- [ ] The memory index (MEMORY.md) is updated to reflect changes

## Source Evidence

- `OpenHarness/src/openharness/memory/` — `add_memory_entry()`, `remove_memory_entry()`
