# ohgo — OpenHarness in Go

A complete Go reimplementation of [OpenHarness](https://github.com/HKUDS/OpenHarness) — the open-source agent harness that wraps an LLM with tool-use, skills, memory, permissions, multi-agent coordination, and MCP support. Distributed as a single static binary.

## Why Go?

The Python OpenHarness works great, but a Go build gives you:

- **Single static binary** — no Python, no virtualenvs, no dependency hell
- **Cross-platform** — `GOOS`/`GOARCH` cross-compilation for Linux, macOS, Windows, ARM
- **Fast startup** — no interpreter warmup, no import resolution at runtime
- **Small footprint** — lean memory usage, suitable for CI and containers

## Quick Start

```bash
# Build
make

# Run
export ANTHROPIC_API_KEY=your_key
./bin/og -p "Explain this codebase"
```

Or install to `$GOPATH/bin`:

```bash
make install
og -p "List all functions in main.go"
```

## Prerequisites

- Go 1.25+
- An LLM API key (Anthropic, OpenAI-compatible, or others)

## Build Commands

```bash
make               # build both binaries (og + ogmo)
make build-og      # build og only
make build-ogmo    # build ogmo only
make test          # run all tests
make test-v        # verbose tests
make test-pkg PKG=./internal/engine  # test a single package
make vet           # go vet
make lint          # golangci-lint
make fmt           # gofmt
make install       # install to GOPATH/bin
make clean         # remove build artifacts
```

## Architecture

```
cmd/og/          # CLI entrypoint (the `og` binary)
cmd/ogmo/        # ohmo personal agent binary
internal/
  engine/        # agent loop, streaming, retry
  tools/         # 28 tool packages (file I/O, shell, search, web, MCP, etc.)
  skills/        # on-demand markdown skill loading
  plugins/       # plugin system (commands, hooks, agents)
  permissions/   # multi-level permission modes and path rules
  hooks/         # PreToolUse/PostToolUse lifecycle hooks
  commands/      # slash commands
  mcp/           # Model Context Protocol client
  memory/        # persistent cross-session memory
  tasks/         # background task lifecycle
  coordinator/   # multi-agent subagent spawning
  prompts/       # system prompt assembly, CLAUDE.md discovery
  config/        # multi-layer config, profile management
  api/           # Anthropic and OpenAI-compatible API clients
  auth/          # authentication flows
  bridge/        # subscription bridges (Claude CLI, Codex CLI)
  sandbox/       # sandboxed command execution
  ui/            # terminal UI
  channels/      # IM channel integrations (Telegram, Slack, Discord, Feishu)
```

### The Agent Loop

The engine drives the core cycle: query → stream response → if `tool_use`, execute tools → append results → loop.

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

## Features

### Tools (28 packages)

| Category | Tools | Description |
|----------|-------|-------------|
| **File I/O** | Bash, Read, Write, Edit, Glob, Grep | Core file operations with permission checks |
| **Search** | WebFetch, WebSearch, LSP | Web and code search |
| **Notebook** | NotebookEdit | Jupyter notebook cell editing |
| **Agent** | Agent, SendMessage, Team | Subagent spawning and coordination |
| **Task** | TaskCreate/Get/List/Update/Stop/Output | Background task management |
| **MCP** | MCPTool, ListMcpResources, ReadMcpResource | Model Context Protocol |
| **Mode** | PlanMode, Worktree | Workflow mode switching |
| **Schedule** | Cron, RemoteTrigger | Scheduled and remote execution |
| **Meta** | Skill, Config, Brief, Sleep, Ask, Todo | Knowledge loading, config, interaction |

Every tool implements the `Tool` interface with JSON Schema for model self-discovery.

### Skills

On-demand knowledge loaded from markdown files. Compatible with the `anthropics/skills` format (YAML frontmatter + body). Drop `.md` files into `~/.openharness/skills/`.

### Plugins

Compatible with the `claude-code/plugins` directory layout. Plugins contribute commands, hooks, and agents via a `plugin.json` manifest.

### Permissions

Multi-level safety with fine-grained control:

| Mode | Behavior | Use Case |
|------|----------|----------|
| **Default** | Ask before write/execute | Daily development |
| **Auto** | Allow everything | Sandboxed environments |
| **Plan Mode** | Block all writes | Review before action |

### Provider Compatibility

Supports the same provider workflows as the Python version:

- **Anthropic-Compatible API** — Claude official, Kimi, GLM, MiniMax
- **Claude Subscription** — local credential bridge
- **OpenAI-Compatible API** — OpenAI, OpenRouter, DashScope, DeepSeek, Groq, Ollama
- **Codex Subscription** — Codex CLI credential bridge
- **GitHub Copilot** — OAuth device flow

## ogmo — Personal Agent

`ogmo` is the Go counterpart to `ohmo`, a personal-agent app with workspace, gateway, and IM channel support:

```bash
ogmo init       # initialize workspace at ~/.ohmo/
ogmo config     # configure provider and channels
ogmo            # run the personal agent
```

## Reference Source

The `OpenHarness/` directory contains a read-only mirror of the upstream Python implementation. It is used as a reference only — do not modify it.

## License

MIT — see [LICENSE](LICENSE).
