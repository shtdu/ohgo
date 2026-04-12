# REQ-EX-007: MCP Server Management

**Pattern:** Event-Driven
**Capability:** Extensibility

## Requirement

When the user adds or removes an MCP server configuration, the system shall persist the configuration change and notify the runtime bridge (per REQ-TL-010).

## Acceptance Criteria

- [ ] MCP servers are added via CLI (`mcp add`) or settings
- [ ] MCP servers are removed via CLI (`mcp remove`)
- [ ] The configuration of external tool servers is persisted and reflected on next session start or MCP reconnection
- [ ] Server connection errors are reported without affecting other tools
- [ ] When adding an MCP server fails due to invalid configuration or connectivity issues, the system reports the error and does not add the server to the configuration

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `mcp` subcommand
- `OpenHarness/src/openharness/mcp/` — client manager
