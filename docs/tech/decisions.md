# Decisions

Technical decision record for ohgo. Each decision includes context, options considered, and rationale.

## ADR-001: Module path

**Decision:** `github.com/shtdu/ohgo`

**Context:** Need a Go module path that identifies the project.

**Options:**
- `github.com/shtdu/ohgo` — project namespace under shtdu org
- `ohgo` — bare name, simpler but can't `go get`

**Rationale:** Standard Go convention requires a hosted path for module resolution.

## ADR-002: Binary names

**Decision:** `og` (agent CLI) and `ogmo` (personal agent)

**Context:** Python version uses `oh` and `ohmo`. Renamed for the Go version.

**Rationale:** Short, memorable, distinct from the Python `oh`. `og` follows the same two-letter pattern.

## ADR-003: CLI framework — cobra

**Decision:** Use `github.com/spf13/cobra` over `urfave/cli` or stdlib `flag`.

**Context:** Need subcommands (`og mcp`, `og plugin`, `og auth`), flag parsing, help generation, shell completions.

**Rationale:** Cobra is the most widely adopted Go CLI framework (kubectl, gh, hugo, talos). Large ecosystem, well-documented, stable. Maps naturally to the Python typer command structure.

## ADR-004: TUI framework — Charm stack

**Decision:** Use `bubbletea` + `lipgloss` + `glamour` + `huh` from the Charm ecosystem.

**Context:** Need interactive TUI with prompts, styled output, markdown rendering.

**Rationale:** Elm-inspired architecture is idiomatic Go. Single ecosystem means consistent APIs and no version conflicts. bubbletea for app loop, huh for prompts/dialogs, glamour for markdown, lipgloss for styling. Replaces textual + prompt_toolkit + rich + questionary from Python.

## ADR-005: Validation — struct tags + go-playground/validator

**Decision:** Use struct tags (`json`, `validate`) with `go-playground/validator/v10`.

**Context:** Python uses pydantic for data validation. Go has no direct equivalent.

**Rationale:** Go's philosophy is struct tags for metadata. `encoding/json` handles serialization. `validator/v10` adds field-level validation rules via tags (`validate:"required,min=1"`). Together they cover pydantic's core use case without reflection-heavy ORM patterns. For JSON Schema generation (tool InputSchema), custom code generates schemas from tagged structs.

## ADR-006: LLM SDKs — both official

**Decision:** Use both `anthropic-sdk-go` and `openai-go`.

**Context:** Python uses both `anthropic` and `openai` SDKs for different providers.

**Rationale:** Each SDK is purpose-built for its provider's API quirks (streaming format, tool use protocol, error handling). Trying to abstract both behind one SDK would leak complexity. The `api.Client` interface normalizes them — the engine doesn't know which provider it's talking to.

## ADR-007: HTTP client — resty

**Decision:** Use `github.com/go-resty/resty/v2` over stdlib `net/http`.

**Context:** Need HTTP client for web fetch, web search, channel integrations, bridge connections.

**Rationale:** resty provides retry, middleware, request/response binding, and a simpler API surface than raw `net/http`. The LLM SDKs handle their own HTTP — resty is for non-LLM HTTP needs.

## ADR-008: WebSocket — gorilla/websocket

**Decision:** Use `github.com/gorilla/websocket`.

**Context:** Needed for Discord gateway and other real-time IM channel connections.

**Rationale:** De facto standard Go WebSocket library. Stable, well-tested, broad adoption. Only used in the channels package.

## ADR-009: MCP — mark3labs/mcp-go

**Decision:** Use `github.com/mark3labs/mcp-go`.

**Context:** Need MCP client to connect to external tool servers via stdio JSON-RPC.

**Rationale:** Most mature Go MCP implementation. Supports client and server modes. Handles stdio transport, JSON-RPC framing, and MCP protocol versioning. Direct replacement for the Python `mcp` package.

## ADR-010: Testing — testify

**Decision:** Use `github.com/stretchr/testify` on top of stdlib `testing`.

**Context:** Need assertions, mocks, and test suites.

**Rationale:** testify is the closest Go equivalent to pytest's ergonomics. `assert`/`require` for assertions, `mock` for interface mocking, `suite` for test grouping. Go's `testing` package is mandatory (test runner) — testify makes tests readable.

## ADR-011: YAML — gopkg.in/yaml.v3

**Decision:** Use `gopkg.in/yaml.v3`.

**Context:** Need YAML for skill frontmatter, plugin manifests, and some config files.

**Rationale:** Standard Go YAML library. Full YAML 1.2 support. Stable API. Used for parsing skill YAML frontmatter and plugin hook definitions.

## ADR-012: Config directory — shared with Python

**Decision:** Use `~/.openharness/` for config, same as the Python version.

**Context:** Users may switch between Python and Go versions.

**Rationale:** Shared config directory means users don't need to reconfigure when switching implementations. `settings.json` format remains compatible. Only the binary changes.

## ADR-013: Package structure — flat internal packages

**Decision:** Each concern gets its own package under `internal/`. Tools get subdirectories per tool.

**Context:** Python uses a deep module hierarchy. Go prefers flat packages.

**Rationale:** Go best practice is small, focused packages. Each `internal/` subdirectory is one package with one job. Tools are an exception — each tool gets its own subdirectory under `internal/tools/` because there are 43+ of them. This avoids single 3000-line files while keeping the tool interface centralized.

## ADR-014: Streaming — normalized event channel

**Decision:** Both Anthropic and OpenAI providers emit into the same `StreamEvent` channel type.

**Context:** Anthropic SSE and OpenAI SSE have different formats and event names.

**Rationale:** The engine shouldn't care about provider differences. Each API client implementation handles its own SSE parsing and emits normalized events. This allows adding new providers without touching the engine.

## ADR-015: Error model — wrapped errors + typed errors

**Decision:** Use `fmt.Errorf("context: %w", err)` for wrapping. Custom error types for programmatic matching.

**Context:** Python uses exception hierarchies. Go uses error values.

**Rationale:** Wrapped errors preserve the chain for `errors.Is()`/`errors.As()`. Custom error types (e.g., `PermissionDeniedError`) allow callers to switch on error kind without string matching. No error code system — Go idiom is type-based matching.
