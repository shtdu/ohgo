# REQ-TL-002: File Operations

**Pattern:** Ubiquitous
**Capability:** Tool Execution

## Requirement

The system shall provide tools for reading, writing, and editing files within the user's workspace.

## Acceptance Criteria

- [ ] Read tool returns file content with line numbers, supporting offset and limit
- [ ] Write tool creates or overwrites files with specified content
- [ ] Edit tool replaces specific text strings in existing files
- [ ] File operations are subject to path permission rules (per Permissions domain)
- [ ] When a file operation fails (not found, permission denied), the tool returns a structured error containing the path and the failure reason

## Source Evidence

- `OpenHarness/src/openharness/tools/file_read_tool.py`
- `OpenHarness/src/openharness/tools/file_write_tool.py`
- `OpenHarness/src/openharness/tools/file_edit_tool.py`
