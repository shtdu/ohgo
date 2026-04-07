# REQ-TL-006: Web Content Fetching

**Pattern:** Event-Driven
**Capability:** Tool Execution

## Requirement

When the agent requests web content, the system shall fetch the specified URL and return the extracted text content.

## Acceptance Criteria

- [ ] Accepts a URL parameter
- [ ] Returns extracted text content, not raw HTML
- [ ] Supports a configurable maximum character limit
- [ ] Handles HTTP errors gracefully with descriptive messages

## Source Evidence

- `OpenHarness/src/openharness/tools/web_fetch_tool.py`
