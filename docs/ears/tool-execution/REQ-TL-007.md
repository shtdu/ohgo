# REQ-TL-007: Web Search

**Pattern:** Event-Driven
**Capability:** Tool Execution

## Requirement

When the agent performs a web search, the system shall query a search engine and return ranked results with titles, URLs, and summaries.

## Acceptance Criteria

- [ ] Accepts a search query string
- [ ] Returns results with title, URL, and summary for each match
- [ ] Supports a configurable maximum number of results
- [ ] Returns results or an error within a configurable timeout period
- [ ] When the search engine query times out, the tool returns a timeout error identifying the search provider

## Source Evidence

- `OpenHarness/src/openharness/tools/web_search_tool.py`
