# REQ-TL-004: File Pattern Search

**Pattern:** Event-Driven
**Capability:** Tool Execution

## Requirement

When the agent searches for files, the system shall match files by glob pattern and return matching paths sorted by modification time.

## Acceptance Criteria

- [ ] Supports standard glob patterns (e.g., `**/*.go`, `src/**/*.ts`)
- [ ] Returns paths sorted by modification time
- [ ] Supports an optional root directory parameter
- [ ] Results are limited to a configurable maximum count
- [ ] When a glob pattern is invalid, the tool returns an error describing the malformed pattern
- [ ] When the root directory is not found or inaccessible, the tool returns an error with the directory path

## Source Evidence

- `OpenHarness/src/openharness/tools/glob_tool.py`
