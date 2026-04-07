# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Goal

This repo is a **complete Go reimplementation** of [OpenHarness](https://github.com/HKUDS/OpenHarness) (originally Python). The goal is a single static binary that delivers the same agent harness — tool-use, skills, memory, permissions, multi-agent coordination, MCP — with easier cross-platform distribution.

## Reference Source

`OpenHarness/` is a **read-only mirror** of the upstream Python implementation. **Do not modify anything inside `OpenHarness/`.** Use it only as a reference when porting features. See `OpenHarness/CLAUDE.md` for the Python architecture details.

## Go Project Structure

This is a Go module. Follow standard Go project layout:

```
cmd/og/          # CLI entrypoint (the `og` binary)
cmd/ogmo/        # ohmo personal agent binary
internal/        # private application packages
  engine/        # agent loop, streaming, retry
  tools/         # 43+ tools (file I/O, shell, search, web, MCP)
  skills/        # on-demand markdown skill loading
  plugins/       # plugin system (commands, hooks, agents)
  permissions/   # multi-level permission modes and path rules
  hooks/         # PreToolUse/PostToolUse lifecycle hooks
  commands/      # slash commands (/help, /commit, /plan, etc.)
  mcp/           # Model Context Protocol client
  memory/        # persistent cross-session memory
  tasks/         # background task lifecycle
  coordinator/   # multi-agent subagent spawning and team coordination
  prompts/       # system prompt assembly, CLAUDE.md discovery
  config/        # multi-layer config, profile management
  api/           # Anthropic and OpenAI-compatible API clients
  auth/          # authentication flows (OAuth device flow, etc.)
  bridge/        # subscription bridges (Claude CLI, Codex CLI)
  ui/            # terminal UI
  channels/      # IM channel integrations (Telegram, Slack, Discord, Feishu)
docs/            # project documentation
go.mod
go.sum
```

## Build & Development Commands

```bash
# Build
go build ./cmd/og

# Run
go run ./cmd/og

# Test all
go test ./...

# Test a single package
go test ./internal/engine/...

# Test a specific test
go test ./internal/engine/ -run TestQueryEngine -v

# Vet
go vet ./...

# Lint (requires golangci-lint)
golangci-lint run

# Format
gofmt -w .
```

## Architecture (Ported from Python)

### Core Agent Loop

The engine drives: query → stream response → if `tool_use`, execute tools → append results → loop. This is the heart of the system and maps directly to `OpenHarness/src/openharness/engine/`.

### Key Patterns to Port

| Python Pattern | Go Equivalent |
|---|---|
| `BaseTool` + Pydantic `input_model` | `Tool` interface with `Execute(ctx, args) (Result, error)` and JSON Schema |
| `ToolRegistry` | Tool registry map with `Register(tool)` / `Get(name)` |
| `QueryEngine` (async) | Struct with methods, using goroutines and channels for streaming |
| `PermissionChecker` | Interface with mode-based implementations |
| `HookExecutor` (PreToolUse/PostToolUse) | Hook chain with registered callbacks |
| `asyncio` throughout | `context.Context` + goroutines + `sync` primitives |
| Pydantic models | Go structs with JSON tags and `encoding/json` or a schema library |
| `anthropic` / `openai` SDK clients | `http.Client` with streaming SSE support against both API formats |

### Data Flow (Same as Python)

```
User Prompt → CLI/TUI → Engine → API Client
                                    ↓ (tool_use)
                              Tool Registry
                                    ↓
                            Permissions + Hooks
                                    ↓
                            Tool Execution
                                    ↓
                              back to Engine
```

### Compatibility Requirements

- **Skills**: Must remain compatible with `anthropics/skills` markdown format (YAML frontmatter + body)
- **Plugins**: Must remain compatible with `claude-code/plugins` directory layout (`plugin.json` manifest)
- **Provider profiles**: Support Anthropic-compatible, OpenAI-compatible, Claude/Codex subscription bridges, and GitHub Copilot
- **Config**: Read from `~/.openharness/` (same config directory as Python version)
- **Permissions**: `settings.json` format compatible with Python version

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use `context.Context` as first param in all public methods that do I/O
- Prefer interfaces for testability (Tool, PermissionChecker, APIClient)
- Error handling: return errors, don't panic. Wrap with `fmt.Errorf("...: %w", err)`
- Keep packages focused — one concern per package under `internal/`
