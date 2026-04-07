# Interfaces

Contract definitions for all core interfaces. These are the seams of the system — every package boundary is defined by an interface.

## Tool (internal/tools)

The fundamental unit of agent capability. Every tool implements this interface.

```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() map[string]any
    Execute(ctx context.Context, args json.RawMessage) (Result, error)
}

type Result struct {
    Content string
    IsError bool
}
```

### Contract

- `Name()` must be unique across the registry — duplicate names panic on registration
- `InputSchema()` returns a valid JSON Schema object (draft-07)
- `Execute()` must be safe for concurrent calls
- `Execute()` must respect context cancellation — return promptly when ctx is done
- `Result.Content` is always populated, even on error (contains error message for the model)
- `Result.IsError` signals tool failure — the model sees the error content and can retry
- Tools must not import engine, permissions, or hooks — they are pure execution units

### Tool Categories

| Category | Read-Only | Examples |
|---|---|---|
| File read | Yes | read, glob, grep |
| File write | No | edit, write |
| Shell | No | bash |
| Search | Yes | grep, web_search |
| Web | Mixed | web_fetch, web_search |
| Agent | No | agent (subagent spawn) |
| MCP | Mixed | mcp_tool (delegated to MCP server) |

### Registry

```go
type Registry struct { ... }

func NewRegistry() *Registry
func (r *Registry) Register(t Tool)     // panics on duplicate name
func (r *Registry) Get(name string) Tool // nil if not found
func (r *Registry) List() []Tool
```

## API Client (internal/api)

Communicates with LLM providers via streaming SSE.

```go
type Client interface {
    Stream(ctx context.Context, opts StreamOptions) (<-chan StreamEvent, error)
}
```

### Contract

- `Stream()` returns immediately — events arrive on the channel asynchronously
- The returned channel is always closed by the provider (EOF) or on error
- Callers must drain the channel or cancel the context to avoid goroutine leaks
- Implementations handle retry with exponential backoff internally — callers see a single logical stream
- API keys come from config — never passed directly to Stream()

### Stream Events

```go
type StreamEvent struct {
    Type string  // "text_delta", "tool_use", "message_complete", "error", "usage"
    Data any     // event-specific payload
}
```

| Event Type | Data Type | Meaning |
|---|---|---|
| `text_delta` | `string` | Incremental text from the model |
| `tool_use` | `ToolCall` | Model requests tool execution |
| `message_complete` | `Message` | Full assistant message assembled |
| `error` | `string` | API error (retry exhausted) |
| `usage` | `UsageStats` | Token usage for billing/tracking |

### Providers

Two concrete implementations:

| Provider | SDK | Purpose |
|---|---|---|
| Anthropic | `anthropic-sdk-go` | Native Claude API |
| OpenAI-compatible | `openai-go` | OpenAI, local models, third-party APIs |

Plus bridge implementations:

| Bridge | Purpose |
|---|---|
| Claude CLI | Proxy via Claude CLI subscription |
| Codex CLI | Proxy via Codex CLI credentials |

## Permission Checker (internal/permissions)

Gate that runs before every tool execution.

```go
type Checker interface {
    Check(ctx context.Context, check Check) (Decision, error)
}

type Decision int  // Allow | Deny | Ask
type Mode string   // "default" | "plan" | "auto"
```

### Contract

- `Check()` must be fast — no network calls, no user interaction
- Return `Ask` to trigger interactive UI prompt
- Return `Deny` to block with a reason
- Return `Allow` to proceed without prompting

### Mode Behavior

| Mode | Read-only tools | Write tools | Shell |
|---|---|---|---|
| `default` | Allow | Ask | Ask |
| `plan` | Allow | Deny | Deny |
| `auto` | Allow | Allow | Allow (with deny-list) |

### Path Rules (from settings.json)

```
permissions:
  allow:
    - "read_file"
    - "glob"
  deny:
    - "bash:rm -rf /"
  paths:
    "/src/**": allow
    "/.env": deny
```

## Hook Executor (internal/hooks)

Lifecycle callbacks that run before and after tool execution.

```go
type Hook func(ctx context.Context, hctx HookContext) (HookResponse, error)

type Executor struct { ... }

func (e *Executor) RegisterPre(h Hook)
func (e *Executor) RegisterPost(h Hook)
func (e *Executor) RunPre(ctx context.Context, hctx HookContext) (HookResponse, error)
func (e *Executor) RunPost(ctx context.Context, hctx HookContext) error
```

### Hook Types

Hooks can execute different backends:

| Type | Mechanism | Use case |
|---|---|---|
| Command | Shell command with `$ARGUMENTS` injection | Safety checks, logging |
| HTTP | POST to endpoint with hook payload | CI integration, audit logging |
| Prompt | Ask model to validate conditions | Deep semantic checks |
| Agent | Full model-based validation | Complex policy enforcement |

### Hook Pattern Matching

Hooks can target specific tools via fnmatch patterns:
```json
{
  "hooks": [
    {"event": "PreToolUse", "pattern": "bash", "command": "check-safety.sh"},
    {"event": "PreToolUse", "pattern": "write_*", "command": "audit-write.sh"}
  ]
}
```

### Contract

- Pre-hooks run in registration order — first block wins
- Post-hooks all run (non-blocking) — errors are logged, not fatal
- `HookResponse.Block = true` stops tool execution immediately
- `HookResponse.ModifiedArgs` replaces the original tool arguments
- Hooks must complete within 30 seconds or context cancellation kills them
- Hooks are registered by plugins via `hooks.json` manifests or by the engine for built-in behavior

## Engine (internal/engine)

The core agent loop. Orchestrates API calls, tool execution, and conversation history.

```go
type Engine struct { ... }

func New(opts Options) *Engine
func (e *Engine) Query(ctx context.Context, prompt string) error
```

### Contract

- `Query()` blocks until the agent loop completes or context is cancelled
- The engine owns conversation history — callers do not mutate it directly
- Max 200 iterations per query (configurable) — hard stop to prevent runaway
- Auto-compaction: when token count exceeds budget, older messages are summarized
  - **Microcompact**: clears old tool results (cheap, no LLM call)
  - **Full compaction**: LLM summarizes older turns into a single message
- Streaming output is sent to the UI via a callback/channel, not return values
- Tool execution modes:
  - **Single tool call**: sequential, event emitted immediately
  - **Multiple tool calls**: concurrent via goroutines, events batched
- Retry with exponential backoff (base 1s, max 30s, jitter) on 429/500/502/503/529

## Config Manager (internal/config)

Multi-layer configuration loader.

```go
type Manager struct { ... }

func NewManager(configDir string) *Manager
func (m *Manager) Load(ctx context.Context) (*Config, error)
```

### Contract

- `Load()` merges all layers (CLI > env > project > user > defaults)
- Later values override earlier ones — last write wins
- Secrets (API keys) are never logged or included in error messages
- Config is immutable after loading — passed by value, not pointer

## Memory Store (internal/memory)

Persistent cross-session memory using markdown files with an index.

```go
type Store struct { ... }

func NewStore(dir string) *Store
func (s *Store) Save(ctx context.Context, key, value string) error
func (s *Store) Load(ctx context.Context, key string) (string, error)
func (s *Store) List(ctx context.Context) ([]MemoryEntry, error)
func (s *Store) Remove(ctx context.Context, name string) error
```

### Storage Layout

```
~/.openharness/memory/<project>/
  MEMORY.md              # index file — one-line pointers to entries
  user_preferences.md    # individual memory entries
  feedback_style.md
  project_context.md
```

### Memory Entry Format

Each entry is a markdown file with YAML frontmatter:

```markdown
---
name: user_preferences
description: User coding preferences
type: user
---

- Prefers table-driven tests
- Uses testify assertions
```

### MEMORY.md Index

```markdown
- [User Preferences](user_preferences.md) — coding style and tooling
- [Feedback Style](feedback_style.md) — response format preferences
```

### Contract

- Max 5 memory files per project (configurable)
- Index file (`MEMORY.md`) max 200 lines
- Each entry is a separate file under the store directory
- Format is compatible with the Python version
- Entries have types: `user`, `feedback`, `project`, `reference`

## Command Registry (internal/commands)

Slash commands for the interactive REPL.

```go
type Command interface {
    Name() string
    Run(ctx context.Context, args string) error
}
```

### Contract

- `Name()` is the command without the `/` prefix (e.g. "help", "commit")
- `Run()` executes synchronously — for long operations, spawn a goroutine internally
- Commands can access the engine to make agent queries
- 54 commands to port from the Python version

## Bridge (internal/bridge)

Subscription proxy that translates API calls to use existing subscriptions.

```go
type Bridge interface {
    Name() string
    Connect(ctx context.Context) error
    Close() error
}
```

## Channel (internal/channels)

IM integration for ogmo.

```go
type Channel interface {
    Name() string
    Connect(ctx context.Context) error
    Close() error
}
```

## UI (internal/ui)

Terminal output and user interaction.

```go
type UI struct { ... }

func New(out io.Writer, in io.Reader) *UI
func (u *UI) Print(msg string)
func (u *UI) Prompt(ctx context.Context, prompt string) (string, error)
```

### Contract

- `Print()` handles markdown rendering and syntax highlighting
- `Prompt()` handles interactive input with history
- UI never blocks the engine — output is buffered
