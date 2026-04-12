# REQ-MC-004: Memory Search

**Pattern:** Event-Driven
**Capability:** Memory and Context

## Requirement

When the agent queries memory, the system shall search memory files by relevance and return matching entries.

## Acceptance Criteria

- [ ] Accepts a search query
- [ ] Returns results ranked by relevance score (text similarity or vector distance depending on the configured backend)
- [ ] Returns file paths and content excerpts
- [ ] Returns an empty result set with no error when no memories match the query
- [ ] When the memory index is unavailable or corrupt, search returns an empty result set with a warning

## Source Evidence

- `OpenHarness/src/openharness/memory/` — `find_relevant_memories()`
