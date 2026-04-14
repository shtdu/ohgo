# Architecture

System architecture for ohgo — the Go reimplementation of OpenHarness.

## Overview

ohgo wraps an LLM into a functional agent with tool-use, skills, memory, permissions, multi-agent coordination, and MCP support. It produces two binaries:

- **og** — the agent CLI (interactive REPL + one-shot mode)
- **ogmo** — the personal agent (IM channel gateway + headless mode)

Both share the same core engine. The difference is the interface layer: og talks to a terminal, ogmo talks to IM channels.

## Architecture Layers

```
┌──────────────────────────────────────────────────────────┐
│                    Interface Layer                        │
│   CLI (og)              TUI (og)         Channels (ogmo) │
└──────────────────────────┬───────────────────────────────┘
                           │
┌──────────────────────────▼───────────────────────────────┐
│                   Command Layer                           │
│   Slash commands ─── /help, /commit, /plan, /mcp, ...    │
└──────────────────────────┬───────────────────────────────┘
                           │
┌──────────────────────────▼───────────────────────────────┐
│                    Engine (Core Loop)                     │
│   Query → Stream → Parse → Execute Tools → Loop          │
│                                                          │
│   ┌─────────────┐  ┌──────────────┐  ┌───────────────┐  │
│   │ API Clients  │  │ Permissions  │  │ Hooks         │  │
│   │ (streaming)  │  │ (gate)       │  │ (lifecycle)   │  │
│   └─────────────┘  └──────────────┘  └───────────────┘  │
└──────────────────────────┬───────────────────────────────┘
                           │
┌──────────────────────────▼───────────────────────────────┐
│                  Capability Layer                         │
│                                                          │
│   ┌──────┐ ┌────────┐ ┌──────┐ ┌─────┐ ┌───────────┐   │
│   │Tools │ │ Skills  │ │ MCP  │ │ Mem │ │ Coordinator│   │
│   └──────┘ └────────┘ └──────┘ └─────┘ └───────────┘   │
│   ┌──────┐ ┌────────┐ ┌──────┐ ┌─────────┐             │
│   │Tasks │ │ Bridge │ │ Auth │ │ Sandbox │              │
│   └──────┘ └────────┘ └──────┘ └─────────┘             │
└──────────────────────────┬───────────────────────────────┘
                           │
┌──────────────────────────▼───────────────────────────────┐
│                  Foundation Layer                         │
│   Config (multi-layer)  ·  Prompts (assembly)            │
│   Plugins (discovery)   ·  UI (terminal output)          │
└──────────────────────────────────────────────────────────┘
```

## Dependency Rules

1. **Dependencies point downward** — higher layers depend on lower layers, never the reverse
2. **Engine is the single orchestrator** — it calls capabilities, nothing calls back into it
3. **API clients are provider-independent** — each client normalizes to the same streaming interface
4. **Tools are self-contained execution units** — they never import engine or permissions
5. **No circular imports** — enforced by Go's package system
6. **UI receives events, never drives the loop** — output channel pattern, not callbacks

## Core Agent Loop

The engine drives a sequential loop:

```
User Prompt
    │
    ▼
 append to conversation history
    │
    ▼
 compact if token budget exceeded
    │
    ▼
 stream from LLM provider (SSE)
    │
    ├─ text response → emit to UI
    │
    └─ tool_use requests → for each:
         │
         ├─ pre-hooks (may block)
         ├─ permission check (allow/deny/ask user)
         ├─ tool execution
         └─ post-hooks
         │
         ▼
    append tool results to history
         │
         ▼
    loop back to compact + stream
```

### Loop Invariants

1. **One stream at a time** — the loop is strictly sequential
2. **Every tool call passes through hooks then permissions** — no shortcuts
3. **Cancellation stops at the next safe point** — between turns, never mid-stream
4. **History grows monotonically** until compaction reclaims space
5. **Hard turn limit** prevents runaway loops

## Agent Loop Variations

### Compaction

When the token budget is exceeded, the engine compacts older messages:

- **Microcompact** — clears old tool result content, keeping recent turns. No LLM call needed.
- **Full compact** — asks the LLM to summarize older turns into a single message. Costs one extra API call.

### Multi-Agent

The coordinator spawns subagents as child processes. Each subagent runs its own engine loop with a scoped tool set. Teams group agents for coordinated workflows.

### MCP Integration

MCP servers are external processes (stdio, SSE, or HTTP) that expose additional tools and resources. The MCP manager handles connection lifecycle and translates MCP calls into the internal tool execution path.

## Configuration

Config is loaded in layers — later layers override earlier ones. Provider profiles abstract connection details so users select a name, not individual fields.

See [config.md](config.md) for the full configuration design.

## Binary Boundaries

### og (agent CLI)

The primary interface. Two modes:
- **Interactive** — REPL with slash commands, streaming output, permission prompts
- **One-shot** — `--prompt` flag, runs a single query, exits

### ogmo (personal agent)

A headless agent that connects to IM channels. Shares the same engine and capabilities, but replaces the terminal interface with channel adapters. Runs persistently, handling messages from multiple users.

## Compatibility Requirements

ohgo must remain compatible with the Python OpenHarness ecosystem:

| Component | Compatibility |
|---|---|
| Skills | YAML frontmatter + markdown body (`anthropics/skills` format) |
| Plugins | `plugin.json` + directory layout (`claude-code/plugins` format) |
| Settings | `settings.json` with the same schema |
| Memory | `MEMORY.md` index + separate files, project memory at `<CWD>/.ohgo/data/memory/` |
| Permissions | `default`, `plan`, `auto` modes with same behavior |
