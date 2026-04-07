# REQ-MC-001: Persistent Cross-Session Memory

**Pattern:** Optional Feature
**Capability:** Memory and Context

## Requirement

Where the memory feature is enabled, the system shall maintain persistent memory entries that survive across sessions, allowing the agent to recall user preferences, project context, and past decisions. When the memory feature is not enabled, the system shall not persist any memory across sessions and shall not load memory files into context.

## Acceptance Criteria

- [ ] Memory content persists across session boundaries and is retrievable by future sessions in the same project
- [ ] Memory entries survive process termination
- [ ] Memory entries are individually addressable and removable

## Source Evidence

- `OpenHarness/src/openharness/memory/` — memory management module
