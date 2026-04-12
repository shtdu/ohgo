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

## ADR-006: LLM SDKs — official Anthropic SDK + raw HTTP for OpenAI

**Decision:** Use `anthropic-sdk-go` for Anthropic, raw `net/http` with custom SSE parsing for OpenAI-compatible and Copilot.

**Context:** Python uses both `anthropic` and `openai` SDKs for different providers.

**Rationale:** The official Anthropic SDK handles Claude-specific streaming, tool use protocol, and error handling well. For OpenAI-compatible providers (including third-party APIs and Copilot), raw HTTP gives more control over SSE parsing and avoids pulling in the large OpenAI Go SDK dependency. The `api.Client` interface normalizes both — the engine doesn't know which provider it's talking to.

## ADR-007: HTTP client — stdlib net/http

**Decision:** Use `net/http` for all non-SDK HTTP needs.

**Context:** Need HTTP client for web fetch, web search, channel integrations, bridge connections.

**Rationale:** stdlib `net/http` is sufficient for the non-LLM HTTP needs. The LLM SDKs handle their own HTTP. No need for a third-party HTTP client library — reduces dependency count.

## ADR-008: WebSocket — deferred

**Decision:** No WebSocket library imported yet. Will evaluate when IM channels are implemented.

**Context:** Needed for Discord gateway and other real-time IM channel connections.

**Rationale:** The channels subsystem is not yet implemented (ogmo is a skeleton). When channel implementations are needed, `github.com/coder/websocket` (formerly `nhooyr.io/websocket`) or `github.com/gorilla/websocket` can be evaluated against the specific requirements.

## ADR-009: MCP — modelcontextprotocol/go-sdk

**Decision:** Use `github.com/modelcontextprotocol/go-sdk`.

**Context:** Need MCP client to connect to external tool servers via stdio, SSE, and streamable HTTP transports.

**Rationale:** Official MCP SDK for Go. Supports client and server modes, multiple transport types (stdio, SSE, streamable HTTP), and handles JSON-RPC framing. Direct replacement for the Python `mcp` package.

## ADR-010: Testing — testify

**Decision:** Use `github.com/stretchr/testify` on top of stdlib `testing`.

**Context:** Need assertions, mocks, and test suites.

**Rationale:** testify is the closest Go equivalent to pytest's ergonomics. `assert`/`require` for assertions, `mock` for interface mocking, `suite` for test grouping. Go's `testing` package is mandatory (test runner) — testify makes tests readable.

## ADR-011: YAML — gopkg.in/yaml.v3

**Decision:** Use `gopkg.in/yaml.v3`.

**Context:** Need YAML for skill frontmatter, plugin manifests, agent definitions, and some config files.

**Rationale:** Standard Go YAML library. Full YAML 1.2 support. Stable API. Used for parsing skill YAML frontmatter, plugin hook definitions, and coordinator agent definitions.

## ADR-012: Config directory — shared with Python

**Decision:** Use `~/.openharness/` for config, same as the Python version.

**Context:** Users may switch between Python and Go versions.

**Rationale:** Shared config directory means users don't need to reconfigure when switching implementations. `settings.json` format remains compatible. Only the binary changes.

## ADR-013: Package structure — flat internal packages

**Decision:** Each concern gets its own package under `internal/`. Tools get subdirectories per tool.

**Context:** Python uses a deep module hierarchy. Go prefers flat packages.

**Rationale:** Go best practice is small, focused packages. Each `internal/` subdirectory is one package with one job. Tools are an exception — each tool gets its own subdirectory under `internal/tools/` because there are 28 of them across 23 packages. This avoids single 3000-line files while keeping the tool interface centralized.

## ADR-014: Streaming — normalized event channel

**Decision:** Both Anthropic and OpenAI providers emit into the same `StreamEvent` channel type.

**Context:** Anthropic SSE and OpenAI SSE have different formats and event names.

**Rationale:** The engine shouldn't care about provider differences. Each API client implementation handles its own SSE parsing and emits normalized events. This allows adding new providers (like Copilot) without touching the engine.

## ADR-015: Error model — wrapped errors + typed errors

**Decision:** Use `fmt.Errorf("context: %w", err)` for wrapping. Custom error types for programmatic matching.

**Context:** Python uses exception hierarchies. Go uses error values.

**Rationale:** Wrapped errors preserve the chain for `errors.Is()`/`errors.As()`. Custom error types (e.g., `PermissionDeniedError`, `RateLimitError`) allow callers to switch on error kind without string matching. No error code system — Go idiom is type-based matching.

## ADR-016: API client registry — factory pattern

**Decision:** Use a factory-based registry mapping `api_format` strings to `ClientFactory` functions.

**Context:** Multiple provider types (Anthropic, OpenAI, Copilot) each need different client construction.

**Rationale:** A factory registry allows the engine to be provider-agnostic. Config specifies `api_format` (e.g., "anthropic", "openai", "copilot"), and the registry produces the correct client. New providers can be added by registering a factory — no engine changes needed.

## ADR-017: Copilot — two-step OAuth

**Decision:** Implement Copilot auth as a two-step flow: GitHub OAuth device code → Copilot API token.

**Context:** GitHub Copilot requires OAuth authentication, not a simple API key.

**Rationale:** The device code flow is user-friendly for CLI apps (no browser redirect needed). The Copilot token is cached with expiry to avoid repeated OAuth flows. Uses `golang.org/x/oauth2` for the device flow implementation.

## ADR-018: Cron scheduling — robfig/cron

**Decision:** Use `github.com/robfig/cron/v3` for scheduled job management.

**Context:** Need cron-like scheduling for recurring agent tasks.

**Rationale:** Most widely used Go cron library. Standard cron expression syntax, timezone support, thread-safe. Provides the `cron_create`, `cron_delete`, `cron_list`, `cron_toggle` tools.

## ADR-019: JSON manipulation — tidwall/gjson + sjson

**Decision:** Use `github.com/tidwall/gjson` and `github.com/tidwall/sjson` for JSON path operations.

**Context:** Need to extract and modify JSON values without full unmarshaling (e.g., tool arguments, API responses).

**Rationale:** gjson/sjson provide fast JSON path queries and mutations without allocating full Go structs. Useful for tool argument extraction and API response handling where full type definitions would be over-engineering.

## ADR-020: Go version — 1.25

**Decision:** Target Go 1.25 (go.mod specifies 1.25.6).

**Context:** Need to choose a minimum Go version for the project.

**Rationale:** Go 1.25 provides the latest language features, standard library improvements, and toolchain updates. As a new project, there's no backwards compatibility constraint — start with the latest stable version.
