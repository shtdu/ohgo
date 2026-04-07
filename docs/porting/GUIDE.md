# Porting Guide

How to port features from the Python OpenHarness to Go. This is the reference for translating Python patterns into idiomatic Go.

## Pattern Translation

### async/await → goroutines + channels

```python
# Python
async def stream_events():
    async for event in client.stream():
        yield event
```

```go
// Go
func streamEvents(ctx context.Context, client Client) <-chan StreamEvent {
    ch := make(chan StreamEvent, 64)
    go func() {
        defer close(ch)
        for event := range client.Stream(ctx) {
            select {
            case ch <- event:
            case <-ctx.Done():
                return
            }
        }
    }()
    return ch
}
```

**Rules:**
- Every goroutine must have an exit condition (context cancellation or channel close)
- Always buffer channels to decouple producer/consumer
- Close channels in the producer, never the consumer
- `select` on `ctx.Done()` in every blocking operation

### Pydantic models → Go structs

```python
# Python
class BashInput(BaseModel):
    command: str = Field(description="The command to run")
    timeout: int = Field(default=120, description="Timeout in seconds")
```

```go
// Go
type BashInput struct {
    Command string `json:"command" validate:"required"`
    Timeout int    `json:"timeout,omitempty"`
}
```

**Rules:**
- Use `json` tags for all API-facing fields
- Use `validate` tags for required fields at system boundaries
- Use `omitempty` for optional fields
- Pointer (`*string`) for truly optional fields where zero value is meaningful
- Generate JSON Schema from struct tags (helper function)

### BaseTool → Tool interface

```python
# Python
class BashTool(BaseTool):
    name = "bash"
    description = "Run a shell command"
    input_model = BashInput

    async def execute(self, args: BashInput, context: ToolContext) -> ToolResult:
        result = subprocess.run(args.command, ...)
        return ToolResult(output=result.stdout)
```

```go
// Go
// internal/tools/bash/bash.go
package bash

type Tool struct{}

func (t *Tool) Name() string        { return "bash" }
func (t *Tool) Description() string { return "Run a shell command" }
func (t *Tool) InputSchema() map[string]any { /* generate from BashInput */ }

func (t *Tool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
    var input BashInput
    if err := json.Unmarshal(args, &input); err != nil {
        return tools.Result{}, fmt.Errorf("parse bash args: %w", err)
    }
    // execute command
    return tools.Result{Content: output}, nil
}
```

**Rules:**
- Unmarshal args at the top of Execute — fail fast on bad input
- Return `tools.Result{Content: msg, IsError: true}` for tool-level errors
- Return Go `error` for infrastructure failures (OOM, context cancelled)
- Always respect `ctx.Done()` in long-running commands

### Tool registration

```python
# Python
registry = ToolRegistry()
registry.register(BashTool())
registry.register(ReadFileTool())
```

```go
// Go
reg := tools.NewRegistry()
reg.Register(&bash.Tool{})
reg.Register(&read.Tool{})
```

### Exception handling → error returns

```python
# Python
try:
    result = tool.execute(args)
except PermissionError:
    return "Permission denied"
except Exception as e:
    return f"Error: {e}"
```

```go
// Go
result, err := tool.Execute(ctx, args)
if err != nil {
    return tools.Result{Content: err.Error(), IsError: true}, nil
}
```

**Key difference:** In Python, errors are exceptions that unwind the stack. In Go, errors are values returned alongside results. Tool execution errors go into `Result{IsError: true}` — the model sees them and can retry. Infrastructure errors (context cancelled, out of memory) return as Go `error` — the engine handles them.

### Configuration loading

```python
# Python (multi-layer)
config = {}
config.update(load_defaults())
config.update(load_user_config())
config.update(load_project_config())
config.update(load_env_vars())
config.update(load_cli_args())
```

```go
// Go
func (m *Manager) Load(ctx context.Context) (*Config, error) {
    cfg := defaults()
    if err := mergeUserConfig(&cfg); err != nil { return nil, err }
    if err := mergeProjectConfig(&cfg); err != nil { return nil, err }
    mergeEnvVars(&cfg)
    // CLI flags merged by caller after Load() returns
    return &cfg, nil
}
```

### Retry with exponential backoff

```python
# Python
for attempt in range(max_retries):
    try:
        return await client.send(request)
    except RateLimitError as e:
        wait = min(base * 2 ** attempt + jitter(), max_wait)
        await asyncio.sleep(wait)
```

```go
// Go
var backoff = backoffConfig{
    Base:   time.Second,
    Max:    60 * time.Second,
    Factor: 2,
    Jitter: true,
}
for attempt := 0; attempt < maxRetries; attempt++ {
    resp, err := client.Send(ctx, req)
    if err == nil { return resp, nil }
    if !isRetryable(err) { return nil, err }
    wait := backoff.Duration(attempt)
    select {
    case <-time.After(wait):
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

## File Mapping

Which Python files map to which Go packages:

| Python | Go |
|---|---|
| `openharness/engine/query.py` | `internal/engine/engine.go` |
| `openharness/engine/messages.py` | `internal/engine/messages.go` |
| `openharness/engine/events.py` | `internal/engine/events.go` |
| `openharness/tools/base.py` | `internal/tools/tool.go` |
| `openharness/tools/registry.py` | `internal/tools/tool.go` (Registry) |
| `openharness/tools/bash.py` | `internal/tools/bash/bash.go` |
| `openharness/tools/read.py` | `internal/tools/read/read.go` |
| `openharness/api/client.py` | `internal/api/anthropic.go` |
| `openharness/api/openai_client.py` | `internal/api/openai.go` |
| `openharness/permissions/checker.py` | `internal/permissions/checker.go` |
| `openharness/hooks/executor.py` | `internal/hooks/hooks.go` |
| `openharness/skills/loader.py` | `internal/skills/skills.go` |
| `openharness/plugins/manager.py` | `internal/plugins/plugins.go` |
| `openharness/config/settings.py` | `internal/config/config.go` |
| `openharness/memory/store.py` | `internal/memory/memory.go` |
| `openharness/mcp/client.py` | `internal/mcp/client.go` |
| `openharness/prompts/assembler.py` | `internal/prompts/prompts.go` |
| `openharness/coordinator/` | `internal/coordinator/coordinator.go` |
| `openharness/tasks/` | `internal/tasks/tasks.go` |
| `openharness/auth/` | `internal/auth/auth.go` |
| `openharness/bridge/` | `internal/bridge/bridge.go` |
| `openharness/ui/` | `internal/ui/ui.go` |
| `openharness/commands/` | `internal/commands/` |
| `openharness/cli.py` | `cmd/og/main.go` |
| `ohmo/cli.py` | `cmd/ogmo/main.go` |
| `openharness/channels/` | `internal/channels/` |

## Porting Order

Recommended order to port features, based on dependencies:

### Phase 1: Foundation
1. `internal/config` — config loading (everything depends on it)
2. `internal/api` — API client interface + Anthropic implementation
3. `internal/tools` — Tool interface + Registry
4. `internal/engine` — core agent loop (stream + tool execution)

### Phase 2: Safety
5. `internal/permissions` — permission checker
6. `internal/hooks` — hook executor
7. `internal/prompts` — system prompt assembly

### Phase 3: Essential Tools (first 10)
8. `internal/tools/read` — read file
9. `internal/tools/write` — write file
10. `internal/tools/edit` — edit file
11. `internal/tools/bash` — shell execution
12. `internal/tools/glob` — file search
13. `internal/tools/grep` — content search
14. `internal/tools/web_fetch` — HTTP fetch
15. `internal/tools/web_search` — web search
16. `internal/tools/agent` — subagent spawn
17. `internal/ui` — terminal output with markdown rendering

### Phase 4: Extended Features
18. `internal/skills` — skill loading
19. `internal/commands` — slash commands
20. `internal/memory` — persistent memory
21. `internal/mcp` — MCP client
22. `internal/plugins` — plugin system

### Phase 5: Advanced
23. `internal/coordinator` — multi-agent
24. `internal/auth` — OAuth flows
25. `internal/bridge` — subscription bridges
26. `internal/tasks` — background tasks
27. `internal/channels` — IM integrations
28. remaining 33 tools
29. remaining slash commands
