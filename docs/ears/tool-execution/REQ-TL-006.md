# REQ-TL-006: Web Content Fetching

**Pattern:** Event-Driven
**Capability:** Tool Execution

## Requirement

When the agent requests web content, the system shall fetch the specified URL and return the extracted text content.

## Acceptance Criteria

- [ ] Accepts a URL parameter
- [ ] Returns extracted text content, not raw HTML
- [ ] Supports a configurable maximum character limit
- [ ] When the requested URL cannot be retrieved, the system returns a descriptive error message indicating the failure reason

## Source Evidence

- `OpenHarness/src/openharness/tools/web_fetch_tool.py`
