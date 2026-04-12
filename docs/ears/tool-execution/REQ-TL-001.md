# REQ-TL-001: Tool Registry

**Pattern:** Complex
**Capability:** Tool Execution

## Requirement

The system shall provide a catalog of available tools that the agent can invoke during execution, each with a defined name, description, input specification, execution behavior.

## Acceptance Criteria

- [ ] Each tool is identified by a unique name
- [ ] Each tool provides a JSON Schema for its input parameters
- [ ] The catalog supports dynamic expansion at runtime (external tool servers, plugins)
- [ ] The agent receives the available tools as part of each API request
- [ ] Tool invocation integrates with the hook system for pre/post execution events (detailed behavior per REQ-EX-005)
- [ ] When a tool schema definition is invalid, the system logs the error and excludes the tool from the catalog

## Source Evidence

- `OpenHarness/src/openharness/tools/` — 43+ tool implementations
- `OpenHarness/src/openharness/tools/base.py` — BaseTool with schema
