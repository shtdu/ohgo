# Memory & Context

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

---

# REQ-MC-002: Memory Discovery on Session Start

**Pattern:** Event-Driven
**Capability:** Memory and Context

## Requirement

When a session starts, the system shall load relevant memory files from the project and user directories into the agent's context.

## Acceptance Criteria

- [ ] Scans project-level memory directories
- [ ] Scans user-level global memory
- [ ] Loads a memory index file (MEMORY.md) summarizing available memories
- [ ] Memory content is available in the system prompt
- [ ] When memory files exist but are unreadable due to permission errors or corruption, the system logs the specific error and skips the affected entries
- [ ] When no memory files are found, the system proceeds with empty context without error

## Source Evidence

- `OpenHarness/src/openharness/memory/` — `scan_memory_files()`, `load_memory_prompt()`

---

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

---

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

---

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

---

# REQ-MC-006: Memory Size Limits

**Pattern:** Event-Driven
**Capability:** Memory and Context

## Requirement

When a memory entry is added, the system shall enforce configurable limits on the number and size of memory entries to prevent unbounded growth.

## Acceptance Criteria

- [ ] Maximum number of memory files is configurable (default: 200 files)
- [ ] Maximum content size per entry is configurable (default: 32KB per entry)
- [ ] When a memory write would exceed a configured limit, the system rejects the write and reports the limit condition to the agent

## Source Evidence

- `OpenHarness/src/openharness/config/settings.py` — `memory.max_files`, `memory.max_entrypoint_lines`

---
