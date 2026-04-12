# REQ-MC-003: Project Instruction Loading

**Pattern:** Event-Driven
**Capability:** Memory and Context

## Requirement

When a session starts, the system shall load project instruction files (CLAUDE.md, AGENTS.md, GEMINI.md) from the project directory hierarchy into the system prompt.

## Acceptance Criteria

- [ ] Discovers instruction files by walking up the directory tree
- [ ] Supports CLAUDE.md, AGENTS.md, and GEMINI.md conventions
- [ ] Merges instructions from multiple hierarchy levels
- [ ] Project instructions are included in the system prompt
- [ ] When no instruction files are found in the project, the system proceeds without error
- [ ] When the project directory is inaccessible or instruction files are unreadable, the system reports the specific error and skips context loading
- [ ] When an instruction file contains invalid or unparsable content, the system logs a warning with the file path and skips that file

## Source Evidence

- `OpenHarness/src/openharness/prompts/claudemd.py` — CLAUDE.md discovery and loading (Go version extends to AGENTS.md, GEMINI.md)
