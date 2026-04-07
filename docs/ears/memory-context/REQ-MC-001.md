# REQ-MC-001: Persistent Cross-Session Memory

**Pattern:** Ubiquitous
**Capability:** Memory and Context

## Requirement

The system shall maintain persistent memory files that survive across sessions, allowing the agent to recall user preferences, project context, and past decisions.

## Acceptance Criteria

- [ ] Memory is stored as markdown files on disk
- [ ] Memory persists after session termination
- [ ] Memory is discoverable by future sessions in the same project

## Source Evidence

- `OpenHarness/src/openharness/memory/` — memory management module
