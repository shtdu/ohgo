# REQ-TL-010: MCP Tool Bridge

**Pattern:** Optional Feature
**Capability:** Tool Execution

## Requirement

Where MCP servers are configured, the system shall discover their available tools, register them in the tool registry, and proxy tool execution requests to the MCP server.

## Acceptance Criteria

- [ ] MCP tools appear alongside built-in tools in the tool registry
- [ ] Tool inputs and outputs are translated between the agent and MCP protocol
- [ ] The system manages MCP server lifecycle (start, connect, disconnect)
- [ ] MCP tool execution respects the same permission system as built-in tools

## Source Evidence

- `OpenHarness/src/openharness/mcp/` — MCP client manager
- `OpenHarness/src/openharness/tools/mcp_tool.py` — McpToolAdapter
