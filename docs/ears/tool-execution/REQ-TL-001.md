# REQ-TL-001: Tool Registry

**Pattern:** Ubiquitous
**Capability:** Tool Execution

## Requirement

The system shall maintain a registry of available tools that the agent can invoke during execution, each with a defined name, description, input schema, and execution handler.

## Acceptance Criteria

- [ ] Tools are registered by name with unique identifiers
- [ ] Each tool provides a JSON Schema for its input parameters
- [ ] The registry supports dynamic registration at runtime (MCP, plugins)
- [ ] The agent receives the tool list as part of each API request

## Source Evidence

- `OpenHarness/src/openharness/tools/` — 43+ tool implementations
- `OpenHarness/src/openharness/tools/base_tool.py` — BaseTool with schema
