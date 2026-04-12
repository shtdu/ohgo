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
- [ ] When content extraction fails despite a successful HTTP response (e.g., empty body, unsupported encoding), the tool returns the raw response with a warning

## Source Evidence

- `OpenHarness/src/openharness/tools/web_fetch_tool.py`
