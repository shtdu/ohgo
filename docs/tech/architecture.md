# Architecture

System architecture for ohgo — the Go reimplementation of OpenHarness.

## Overview

ohgo is a single static binary that wraps an LLM into a functional agent. It provides tool-use, skills, memory, permissions, multi-agent coordination, and MCP support. Two binaries are produced:

- **og** — the agent CLI (interactive REPL + one-shot mode)
- **ogmo** — the personal agent (IM channel gateway + headless mode)

## Package Dependency Graph

```
cmd/og ─────────────────────────────────────────────────────┐
cmd/ogmo ───────────────────────────────────────────────────┤
                                                            │
  ┌─────────┐                                               │
  │ engine   │ ← core agent loop                            │
  └────┬────┘                                               │
       │ depends on                                          │
       ├──── api/        (LLM provider clients)              │
       ├──── tools/      (Tool interface + registry)         │
       ├──── permissions/ (pre-execution permission check)   │
       ├──── hooks/      (pre/post tool lifecycle)           │
       └──── config/     (merged config)                     │
                                                            │
  ┌─────────┐                                               │
  │ prompts  │ ← system prompt assembly                     │
  └────┬────┘                                               │
       ├──── skills/     (markdown skill loading)            │
       └──── config/     (CLAUDE.md discovery)               │
                                                            │
  ┌─────────┐                                               │
  │ commands │ ← slash commands (/help, /commit, etc.)      │
  └────┬────┘                                               │
       └──── engine/      (can invoke the agent loop)        │
                                                            │
  Standalone packages (no cross-dependencies):               │
  ┌──────────────┐  ┌──────────┐  ┌──────────┐             │
  │ coordinator  │  │ memory   │  │ tasks    │              │
  └──────────────┘  └──────────┘  └──────────┘             │
  ┌──────────────┐  ┌──────────┐  ┌──────────┐             │
  │ mcp          │  │ auth     │  │ bridge   │              │
  └──────────────┘  └──────────┘  └──────────┘             │
  ┌──────────────┐  ┌──────────┐                            │
  │ plugins      │  │ channels │                            │
  └──────────────┘  └──────────┘                            │
  ┌──────────────┐                                          │
  │ ui           │                                          │
  └──────────────┘                                          │
                                                            │
  └──────────────────────────────────────────────────────────┘
```

## Dependency Rules

1. **engine/** depends on api, tools, permissions, hooks, config — nothing else
2. **tools/** has zero internal dependencies — each tool is self-contained
3. **api/** has zero internal dependencies — provider clients are independent
4. **No circular imports** — enforced by Go's package system
5. **ui/** never imports engine — UI receives events, doesn't drive the loop

## Data Flow

```
User Prompt
    │
    ▼
 CLI (cobra) ─── parse flags, load config
    │
    ▼
 Engine.Query(prompt)
    │
    ├─ 1. Prompts.Assembler.Build() → system prompt
    │
    ├─ 2. api.Client.Stream(messages, tools, system)
    │       │
    │       ▼  (SSE stream)
    │   StreamEvent channel
    │       │
    │       ├─ text_delta → UI output
    │       │
    │       └─ message_complete with tool_use
    │            │
    │            ▼
    │       3. For each tool call:
    │            │
    │            ├─ hooks.Executor.RunPre()
    │            │     └─ block? → stop, report reason
    │            │
    │            ├─ permissions.Checker.Check()
    │            │     └─ deny?  → stop
    │            │     └─ ask?   → UI prompt user
    │            │
    │            ├─ tools.Registry.Get(name).Execute(args)
    │            │
    │            └─ hooks.Executor.RunPost()
    │
    ├─ 4. Append tool results to messages
    │
    └─ 5. Loop back to step 2 (until no more tool_use or max turns)
```

## Core Agent Loop Invariants

1. **At most one API stream is active at a time** — the loop is sequential
2. **Tool execution is sequential by default** — parallel only when explicitly configured
3. **Every tool call goes through permissions + hooks** — no bypass
4. **Context cancellation stops the loop at the next safe point** — between API calls or between tool executions, never mid-stream
5. **Conversation history grows monotonically** until compaction triggers (token budget exceeded)
6. **Max turn limit prevents infinite loops** — configurable, default 200

## Binary Structure

```
cmd/og/main.go     → og binary
  ├── cobra root command (interactive mode)
  ├── --prompt flag (one-shot mode)
  ├── --model flag (override config)
  ├── --permission flag (default|plan|auto)
  └── subcommands: mcp, plugin, auth, provider

cmd/ogmo/main.go   → ogmo binary
  ├── cobra root command (headless agent)
  ├── --channel flag (telegram|slack|discord|feishu)
  └── workspace at ~/.ohmo/
```

## Configuration Layers (highest precedence first)

| Priority | Source | Location |
|---|---|---|
| 1 | CLI flags | `--model`, `--permission`, etc. |
| 2 | Environment variables | `OPENHARNESS_MODEL`, `ANTHROPIC_API_KEY` |
| 3 | Project config | `./.openharness/settings.json` |
| 4 | User config | `~/.openharness/settings.json` |
| 5 | Defaults | hardcoded in config package |

## Storage Paths

| Path | Purpose |
|---|---|
| `~/.openharness/` | User config directory (shared with Python version) |
| `~/.openharness/settings.json` | Permission rules, profiles, hooks |
| `~/.openharness/data/memory/` | Cross-session memory store |
| `./.openharness/` | Project-level config |
| `./CLAUDE.md` | Project-level system prompt injection |
| `~/.openharness/CLAUDE.md` | User-level system prompt injection |

## Compatibility Requirements

These must remain compatible with the Python OpenHarness:

| Component | Format | Compatibility |
|---|---|---|
| Skills | YAML frontmatter + markdown body | `anthropics/skills` format |
| Plugins | `plugin.json` + directory layout | `claude-code/plugins` format |
| Settings | `settings.json` | Same schema as Python version |
| Memory | `MEMORY.md` index + separate files | Same directory structure |
| Permission modes | `default`, `plan`, `auto` | Same behavior as Python version |
