# REQ-TL-011: Tool Discovery

**Pattern:** Event-Driven
**Capability:** Tool Execution

## Requirement

When the agent queries available tools, the system shall search tool names and descriptions and return matching results.

## Acceptance Criteria

- [ ] Accepts a search query string
- [ ] Searches across tool names and descriptions
- [ ] Returns matching tools with their descriptions and parameter schemas
- [ ] When a tool search returns no results, the system returns an empty list without error
- [ ] When the search query is empty or exceeds a maximum length, the tool returns a validation error

## Source Evidence

- `OpenHarness/src/openharness/tools/tool_search_tool.py`
