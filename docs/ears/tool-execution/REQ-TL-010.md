# REQ-TL-010: MCP Tool Bridge

**Pattern:** Optional Feature
**Capability:** Tool Execution

## Requirement

Where external tool servers are configured, the system shall discover their available tools, make them available alongside built-in tools, and relay execution requests and results between the agent and the external server.

## Acceptance Criteria

- [ ] External tools appear alongside built-in tools in the tool catalog
- [ ] External tools accept the same input format and return results in the same output format as built-in tools, regardless of the external server's native format
- [ ] The system manages external server connections (start, connect, disconnect)
- [ ] External tool execution respects the same permission system as built-in tools

## Source Evidence

- `OpenHarness/src/openharness/mcp/` — MCP client manager
- `OpenHarness/src/openharness/tools/mcp_tool.py` — McpToolAdapter
