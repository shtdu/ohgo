# REQ-TL-001: Tool Registry

**Pattern:** Ubiquitous
**Capability:** Tool Execution

## Requirement

The system shall provide a catalog of available tools that the agent can invoke during execution, each with a defined name, description, input specification, and execution behavior.

## Acceptance Criteria

- [ ] Each tool is identified by a unique name
- [ ] Each tool provides a JSON Schema for its input parameters
- [ ] The catalog supports dynamic expansion at runtime (external tool servers, plugins)
- [ ] The agent receives the available tools as part of each API request
- [ ] Before and after each tool invocation, lifecycle callbacks execute and may alter or prevent the operation (per REQ-EX-005)

## Source Evidence

- `OpenHarness/src/openharness/tools/` — 43+ tool implementations
- `OpenHarness/src/openharness/tools/base_tool.py` — BaseTool with schema
