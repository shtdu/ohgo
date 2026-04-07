# REQ-EX-007: MCP Server Management

**Pattern:** Event-Driven
**Capability:** Extensibility

## Requirement

When the user adds or removes an MCP server configuration, the system shall update the MCP client connections and reflect the change in the tool registry.

## Acceptance Criteria

- [ ] MCP servers are added via CLI (`mcp add`) or settings
- [ ] MCP servers are removed via CLI (`mcp remove`)
- [ ] Tool registry updates to include or exclude MCP tools accordingly
- [ ] Server connection errors are reported without crashing

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `mcp` subcommand
- `OpenHarness/src/openharness/mcp/` — client manager
