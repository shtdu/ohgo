# REQ-MC-003: Project Instruction Loading

**Pattern:** Event-Driven
**Capability:** Memory and Context

## Requirement

When a session starts, the system shall discover and load project instruction files (CLAUDE.md, AGENTS.md, GEMINI.md) from the project directory hierarchy.

## Acceptance Criteria

- [ ] Discovers instruction files by walking up the directory tree
- [ ] Supports CLAUDE.md, AGENTS.md, and GEMINI.md conventions
- [ ] Merges instructions from multiple hierarchy levels
- [ ] Project instructions are included in the system prompt

## Source Evidence

- `OpenHarness/src/openharness/prompts/` — CLAUDE.md discovery and loading
