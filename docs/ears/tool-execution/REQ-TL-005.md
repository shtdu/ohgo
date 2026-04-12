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
- [ ] When a regex pattern is invalid, the tool returns a parse error identifying the offending portion of the pattern
- [ ] When file read errors occur (permission denied, binary file), the tool returns an error identifying the affected file

## Source Evidence

- `OpenHarness/src/openharness/tools/grep_tool.py`
