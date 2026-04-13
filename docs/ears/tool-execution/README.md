# Tool Execution

How tools are registered, discovered, invoked, sandboxed, and monitored — the complete tool lifecycle.

## Requirements

| ID | Title | Pattern |
|----|-------|---------|
| [REQ-TL-001](#req-tl-001-tool-registry) | Tool Registry | Complex |
| [REQ-TL-002](#req-tl-002-file-operations) | File Operations | Ubiquitous |
| [REQ-TL-003](#req-tl-003-shell-command-execution) | Shell Command Execution | Event-Driven |
| [REQ-TL-004](#req-tl-004-file-pattern-search) | File Pattern Search | Event-Driven |
| [REQ-TL-005](#req-tl-005-content-search) | Content Search | Event-Driven |
| [REQ-TL-006](#req-tl-006-web-content-fetching) | Web Content Fetching | Event-Driven |
| [REQ-TL-007](#req-tl-007-web-search) | Web Search | Event-Driven |
| [REQ-TL-008](#req-tl-008-code-intelligence) | Code Intelligence | Optional Feature |
| [REQ-TL-009](#req-tl-009-notebook-cell-editing) | Notebook Cell Editing | Optional Feature |
| [REQ-TL-010](#req-tl-010-mcp-tool-bridge) | MCP Tool Bridge | Complex |
| [REQ-TL-011](#req-tl-011-tool-discovery) | Tool Discovery | Event-Driven |

## Dependencies

Cross-references to other domains:
- [Extensibility](../extensibility/README.md)

## Details

## REQ-TL-001: Tool Registry

**Pattern:** Complex

### Requirement

The system shall provide a catalog of available tools that the agent can invoke during execution, each with a defined name, description, input specification, execution behavior.

### Acceptance Criteria

- [ ] Each tool is identified by a unique name
- [ ] Each tool provides a JSON Schema for its input parameters
- [ ] The catalog supports dynamic expansion at runtime (external tool servers, plugins)
- [ ] The agent receives the available tools as part of each API request
- [ ] Tool invocation integrates with the hook system for pre/post execution events (detailed behavior per REQ-EX-005)
- [ ] When a tool schema definition is invalid, the system logs the error and excludes the tool from the catalog

### Source Evidence

- `OpenHarness/src/openharness/tools/` — 43+ tool implementations
- `OpenHarness/src/openharness/tools/base.py` — BaseTool with schema


---

## REQ-TL-002: File Operations

**Pattern:** Ubiquitous

### Requirement

The system shall provide tools for reading, writing, and editing files within the user's workspace.

### Acceptance Criteria

- [ ] Read tool returns file content with line numbers, supporting offset and limit
- [ ] Write tool creates or overwrites files with specified content
- [ ] Edit tool replaces specific text strings in existing files
- [ ] File operations are subject to path permission rules (per Permissions domain)
- [ ] When a file operation fails (not found, permission denied), the tool returns a structured error containing the path and the failure reason

### Source Evidence

- `OpenHarness/src/openharness/tools/file_read_tool.py`
- `OpenHarness/src/openharness/tools/file_write_tool.py`
- `OpenHarness/src/openharness/tools/file_edit_tool.py`


---

## REQ-TL-003: Shell Command Execution

**Pattern:** Event-Driven

### Requirement

When the agent invokes the command execution tool, the system shall execute the specified command in the configured working directory and return captured output.

### Acceptance Criteria

- [ ] Commands execute in the configured working directory
- [ ] Captures both standard output and standard error output
- [ ] Partial output captured before timeout is included in the result
- [ ] A default timeout applies when no explicit timeout is specified; exceeding it terminates the command and returns a timeout error message
- [ ] The working directory persists between sequential command invocations
- [ ] When a shell command returns a non-zero exit code, the tool returns both stdout and stderr to the agent
- [ ] When the command executable is not found or execution is denied by permissions, the tool returns an error with the command name and failure reason

### Source Evidence

- `OpenHarness/src/openharness/tools/bash_tool.py`


---

## REQ-TL-004: File Pattern Search

**Pattern:** Event-Driven

### Requirement

When the agent searches for files, the system shall match files by glob pattern and return matching paths sorted by modification time.

### Acceptance Criteria

- [ ] Supports standard glob patterns (e.g., `**/*.go`, `src/**/*.ts`)
- [ ] Returns paths sorted by modification time
- [ ] Supports an optional root directory parameter
- [ ] Results are limited to a configurable maximum count
- [ ] When a glob pattern is invalid, the tool returns an error describing the malformed pattern
- [ ] When the root directory is not found or inaccessible, the tool returns an error with the directory path

### Source Evidence

- `OpenHarness/src/openharness/tools/glob_tool.py`


---

## REQ-TL-005: Content Search

**Pattern:** Event-Driven

### Requirement

When the agent searches file contents, the system shall match lines by regular expression pattern and return matching results with context lines.

### Acceptance Criteria

- [ ] Supports full regex syntax
- [ ] Returns matching lines with configurable context (before/after lines)
- [ ] Supports file type filtering by glob pattern
- [ ] Supports case-insensitive search mode
- [ ] When a regex pattern is invalid, the tool returns a parse error identifying the offending portion of the pattern
- [ ] When file read errors occur (permission denied, binary file), the tool returns an error identifying the affected file

### Source Evidence

- `OpenHarness/src/openharness/tools/grep_tool.py`


---

## REQ-TL-006: Web Content Fetching

**Pattern:** Event-Driven

### Requirement

When the agent requests web content, the system shall fetch the specified URL and return the extracted text content.

### Acceptance Criteria

- [ ] Accepts a URL parameter
- [ ] Returns extracted text content, not raw HTML
- [ ] Supports a configurable maximum character limit
- [ ] When the requested URL cannot be retrieved, the system returns a descriptive error message indicating the failure reason
- [ ] When content extraction fails despite a successful HTTP response (e.g., empty body, unsupported encoding), the tool returns the raw response with a warning

### Source Evidence

- `OpenHarness/src/openharness/tools/web_fetch_tool.py`


---

## REQ-TL-007: Web Search

**Pattern:** Event-Driven

### Requirement

When the agent performs a web search, the system shall query a search engine and return ranked results with titles, URLs, and summaries.

### Acceptance Criteria

- [ ] Accepts a search query string
- [ ] Returns results with title, URL, and summary for each match
- [ ] Supports a configurable maximum number of results
- [ ] Returns results or an error within a configurable timeout period
- [ ] When the search engine query times out, the tool returns a timeout error identifying the search provider

### Source Evidence

- `OpenHarness/src/openharness/tools/web_search_tool.py`


---

## REQ-TL-008: Code Intelligence

**Pattern:** Optional Feature

### Requirement

Where a language intelligence service is available for the file type, the system shall provide code intelligence and navigation operations.

### Acceptance Criteria

- [ ] Supports operations: goToDefinition, findReferences, hover, documentSymbol, workspaceSymbol
- [ ] Results include the source file path and line number for each symbol
- [ ] Returns structured results suitable for agent interpretation
- [ ] When no language intelligence service is available for the target file type, the system returns a descriptive error

### Source Evidence

- `OpenHarness/src/openharness/tools/lsp_tool.py`


---

## REQ-TL-009: Notebook Cell Editing

**Pattern:** Optional Feature

### Requirement

Where a Jupyter notebook file is targeted, the system shall provide cell-level editing operations including replacing, inserting, and deleting cells.

### Acceptance Criteria

- [ ] Supports cell types: code and markdown
- [ ] Supports edit modes: replace, insert, delete
- [ ] Operates on individual cells by index
- [ ] Preserves notebook structure and metadata
- [ ] When the cell index is out of bounds or the notebook format is invalid, the tool returns an error identifying the issue

### Source Evidence

- `OpenHarness/src/openharness/tools/notebook_edit_tool.py`


---

## REQ-TL-010: MCP Tool Bridge

**Pattern:** Complex

### Requirement

Where external tool servers are configured, the system shall discover their tools, expose them alongside built-in tools, relaying execution bidirectionally with the external server.

### Acceptance Criteria

- [ ] External tools appear alongside built-in tools in the tool catalog
- [ ] External tools accept the same input format and return results in the same output format as built-in tools, regardless of the external server's native format
- [ ] The system manages external server connections (connect to running servers, disconnect on shutdown)
- [ ] External tool execution respects the same permission system as built-in tools
- [ ] When an MCP server connection fails or times out, the tool returns a connection error containing the server name and failure reason
- [ ] When invalid input is sent to an external tool, the system returns the external server's error response to the agent

### Source Evidence

- `OpenHarness/src/openharness/mcp/` — MCP client manager
- `OpenHarness/src/openharness/tools/mcp_tool.py` — McpToolAdapter


---

## REQ-TL-011: Tool Discovery

**Pattern:** Event-Driven

### Requirement

When the agent queries available tools, the system shall search tool names and descriptions and return matching results.

### Acceptance Criteria

- [ ] Accepts a search query string
- [ ] Searches across tool names and descriptions
- [ ] Returns matching tools with their descriptions and parameter schemas
- [ ] When a tool search returns no results, the system returns an empty list without error
- [ ] When the search query is empty or exceeds a maximum length, the tool returns a validation error

### Source Evidence

- `OpenHarness/src/openharness/tools/tool_search_tool.py`
