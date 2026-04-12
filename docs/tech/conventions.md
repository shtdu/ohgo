# Conventions

Coding standards, naming rules, and patterns for the ohgo codebase.

## Go Style

- Run `gofmt -w .` before every commit — no style debates
- Run `go vet ./...` — must pass clean
- Run `golangci-lint run` — must pass clean
- Follow [Effective Go](https://go.dev/doc/effective_go) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

## Naming

### Packages

- Lowercase, single word, no underscores: `tools`, `engine`, `mcp`
- Package name should not repeat at the type level: `tools.Tool` not `tools.ToolTool`
- Avoid `util`, `common`, `base` — name by purpose

### Types

- PascalCase for exported, camelCase for unexported
- Interface names: verb+er (`Checker`, `Executor`, `Loader`) or noun (`Client`, `Registry`)
- Struct names: noun (`Engine`, `Config`, `Profile`)
- Error types: `Err` prefix or `Error` suffix (`ErrNotFound`, `PermissionError`)

### Methods

- PascalCase exported, camelCase unexported
- Boolean getters: `IsReadOnly()`, not `GetReadOnly()`
- Constructor: `New()`, `NewXxx()` — no `Create` or `Make`

### Variables

- Short names in tight scope: `r` for reader, `w` for writer
- Descriptive names in wide scope: `permissionChecker`, not `pc`
- Acronyms are all-caps: `APIKey`, `HTTPClient`, `JSONSchema`
- `ctx` for `context.Context` — always, no exceptions

## Error Handling

```go
// Wrap errors with context — always include what you were doing
file, err := os.Open(path)
if err != nil {
    return fmt.Errorf("open config file %s: %w", path, err)
}

// Sentinel errors for public API
var ErrNotFound = errors.New("tool not found")

// Custom error types for programmatic matching
type PermissionDeniedError struct {
    Tool   string
    Reason string
}
func (e *PermissionDeniedError) Error() string {
    return fmt.Sprintf("permission denied for tool %q: %s", e.Tool, e.Reason)
}

// Match with errors.Is / errors.As
var permErr *PermissionDeniedError
if errors.As(err, &permErr) {
    // handle permission error
}
```

### Rules

- Never panic in library code — only in main/init for unrecoverable startup errors
- Never swallow errors with `_ = mightFail()` — handle, wrap, or return
- Error messages are lowercase, no trailing period: `"open config: %w"`
- Wrap at the boundary where context is meaningful — don't wrap in every function

## Context

```go
// context.Context is always the first parameter
func (e *Engine) Query(ctx context.Context, prompt string) error

// Never store a context in a struct — pass it explicitly
// BAD:
type Engine struct { ctx context.Context }
// GOOD:
func (e *Engine) Query(ctx context.Context, ...)

// Derive contexts with timeouts for external calls
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
```

## Struct Tags

```go
type Profile struct {
    Name     string `json:"name"     validate:"required"`
    Provider string `json:"provider" validate:"required,oneof=anthropic openai copilot codex"`
    BaseURL  string `json:"base_url,omitempty"`
    APIKey   string `json:"api_key,omitempty"`
    Model    string `json:"model"    validate:"required"`
}
```

- Always include `json` tags for types that cross API boundaries
- Use `omitempty` for optional fields
- Use `validate` tags for input validation at system boundaries
- Never validate internal data — trust your own code

## Testing

```go
package tools_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestRegistryGet(t *testing.T) {
    reg := NewRegistry()
    tool := &mockTool{name: "bash"}

    reg.Register(tool)

    got := reg.Get("bash")
    assert.Equal(t, tool, got)
    assert.Nil(t, reg.Get("nonexistent"))
}

func TestRegistryDuplicatePanics(t *testing.T) {
    reg := NewRegistry()
    reg.Register(&mockTool{name: "bash"})

    assert.Panics(t, func() {
        reg.Register(&mockTool{name: "bash"})
    })
}
```

### Rules

- Test files: `foo_test.go` in the same package (white-box testing is the norm in this project)
- Use `require` for setup that must succeed (test can't continue without it)
- Use `assert` for assertions that should be true (test continues on failure)
- Table-driven tests for multi-case scenarios
- Mock interfaces with testify/mock — don't mock concrete types
- Test file layout mirrors source: `internal/tools/tool_test.go`

### Test Organization

```
internal/
  engine/
    engine.go
    engine_test.go          # unit tests
    testdata/               # test fixtures
  tools/
    tool.go
    tool_test.go
    bash/
      bash.go
      bash_test.go
    read/
      read.go
      read_test.go
```

## Concurrency

```go
// Pass context, not channels, across API boundaries
func (e *Engine) Query(ctx context.Context, prompt string) error {

// Use channels for producer-consumer patterns
events := make(chan StreamEvent, 64) // buffered to decouple

// Use sync.Mutex for shared state
type Registry struct {
    mu    sync.RWMutex
    tools map[string]Tool
}

// Prefer RWMutex when reads dominate writes
func (r *Registry) Get(name string) Tool {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return r.tools[name]
}
```

### Rules

- Document which goroutine owns which channel
- Always select on ctx.Done() in goroutines that might block
- Close channels in the producer, never the consumer
- Use `sync.Once` for lazy initialization
- Never start a goroutine without a way to stop it

## Logging

```go
import "log/slog"

// Structured logging — no printf-style
slog.Info("tool executed",
    "tool", name,
    "duration_ms", elapsed.Milliseconds(),
    "error", err,
)

// Levels:
//   slog.Debug — verbose, dev-only (off by default)
//   slog.Info  — normal operation
//   slog.Warn  — unexpected but recoverable
//   slog.Error — operation failed
```

## File Organization

```
internal/tools/
  tool.go              # interface + registry
  bash/
    bash.go            # bash tool implementation
  read/
    read.go            # read tool implementation
  write/
    write.go           # write tool implementation
```

- One file per tool (or per closely-related tool group)
- Interface in the parent package, implementations in subdirectories
- Avoid files > 500 lines — split by concern, not by type

## Import Ordering

```go
import (
    // stdlib
    "context"
    "encoding/json"
    "fmt"

    // external
    "github.com/charmbracelet/lipgloss"
    "github.com/stretchr/testify/assert"

    // internal
    "github.com/shtdu/ohgo/internal/api"
    "github.com/shtdu/ohgo/internal/tools"
)
```

Three groups, separated by blank lines: stdlib, external, internal.
