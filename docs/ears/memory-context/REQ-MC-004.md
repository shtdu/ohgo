# REQ-MC-004: Memory Search

**Pattern:** Event-Driven
**Capability:** Memory and Context

## Requirement

When the agent queries memory, the system shall search memory files by relevance and return matching entries.

## Acceptance Criteria

- [ ] Accepts a search query
- [ ] Returns ranked results by relevance
- [ ] Returns file paths and content excerpts

## Source Evidence

- `OpenHarness/src/openharness/memory/` — `find_relevant_memories()`
