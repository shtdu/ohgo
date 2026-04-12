# REQ-TL-010: MCP Tool Bridge

**Pattern:** Complex
**Capability:** Tool Execution

## Requirement

Where external tool servers are configured, the system shall discover their tools, expose them alongside built-in tools, relaying execution bidirectionally with the external server.

## Acceptance Criteria

- [ ] External tools appear alongside built-in tools in the tool catalog
- [ ] External tools accept the same input format and return results in the same output format as built-in tools, regardless of the external server's native format
- [ ] The system manages external server connections (connect to running servers, disconnect on shutdown)
- [ ] External tool execution respects the same permission system as built-in tools
- [ ] When an MCP server connection fails or times out, the tool returns a connection error containing the server name and failure reason
- [ ] When invalid input is sent to an external tool, the system returns the external server's error response to the agent

## Source Evidence

- `OpenHarness/src/openharness/mcp/` — MCP client manager
- `OpenHarness/src/openharness/tools/mcp_tool.py` — McpToolAdapter
