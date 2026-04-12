# REQ-MC-001: Persistent Cross-Session Memory

**Pattern:** Optional Feature
**Capability:** Memory and Context

## Requirement

Where the memory feature is enabled, the system shall maintain persistent memory entries that survive across sessions, enabling recall of user preferences, project context, past decisions. When disabled, no memory persists across sessions; no memory files load into context.

## Acceptance Criteria

- [ ] Memory content persists across session boundaries and is retrievable by future sessions in the same project
- [ ] Memory entries survive process termination
- [ ] Memory entries are individually addressable and removable
- [ ] When the memory store is corrupted or unreadable, the system logs the error and continues without memory

## Source Evidence

- `OpenHarness/src/openharness/memory/` — memory management module
