# REQ-TL-005: Content Search

**Pattern:** Event-Driven
**Capability:** Tool Execution

## Requirement

When the agent searches file contents, the system shall match lines by regular expression pattern and return matching results with context lines.

## Acceptance Criteria

- [ ] Supports full regex syntax
- [ ] Returns matching lines with configurable context (before/after lines)
- [ ] Supports file type filtering by glob pattern
- [ ] Supports case-insensitive search mode

## Source Evidence

- `OpenHarness/src/openharness/tools/grep_tool.py`
