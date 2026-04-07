# REQ-MC-004: Memory Search

**Pattern:** Event-Driven
**Capability:** Memory and Context

## Requirement

When the agent queries memory, the system shall search memory files by relevance and return matching entries.

## Acceptance Criteria

- [ ] Accepts a search query
- [ ] Returns results ranked by text similarity to the query
- [ ] Returns file paths and content excerpts
- [ ] Returns an empty result set with no error when no memories match the query

## Source Evidence

- `OpenHarness/src/openharness/memory/` — `find_relevant_memories()`
