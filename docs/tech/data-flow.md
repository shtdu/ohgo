# Data Flow

How data moves through the system — from user input to LLM response and back.

## Request Lifecycle

### 1. Input → Engine

```
User types prompt in REPL or passes --prompt flag
    │
    ▼
cobra command handler
    │ resolves flags, loads config, creates Engine
    ▼
Engine.Query(ctx, prompt)
```

### 2. Engine → API Client

```
Engine.Query()
    │
    ├─ Prompts.Assembler.Build(ctx)
    │     ├─ reads ~/.openharness/CLAUDE.md (user global)
    │     ├─ reads ./CLAUDE.md (project)
    │     ├─ loads active skills
    │     └─ composes system prompt string
    │
    ├─ builds []api.Message from:
    │     ├─ system prompt
    │     ├─ conversation history (previous turns)
    │     └─ new user message
    │
    ├─ builds []api.ToolDef from:
    │     └─ tools.Registry.List() → Name, Description, InputSchema
    │
    └─ api.Client.Stream(ctx, StreamOptions{...})
          │
          ▼
      returns <-chan StreamEvent
```

### 3. Stream Processing Loop

```
for event := range eventChan {
    switch event.Type {

    case "text_delta":
        │
        ▼
        UI.Print(delta)                    // incremental output to terminal
        append to current assistant message

    case "tool_use":
        │
        ▼
        collect ToolCall{ID, Name, Input}
        (may receive multiple tool_use blocks in one message)

    case "message_complete":
        │
        ▼
        if tool calls pending:
            execute tools (see Tool Execution below)
        else:
            done — return to caller

    case "error":
        return error to caller

    case "usage":
        track token counts for compaction decision
    }
}
```

### 4. Tool Execution

```
For each ToolCall in message:
    │
    ├─ hooks.Executor.RunPre(ctx, HookContext{ToolName, Args})
    │     │
    │     └─ if Block: return HookResponse{Block: true, Reason: "..."}
    │        (tool not executed, reason sent back to model)
    │
    ├─ permissions.Checker.Check(ctx, Check{ToolName, Args})
    │     │
    │     ├─ Allow → proceed
    │     ├─ Deny  → skip tool, send denial to model
    │     └─ Ask   → UI prompts user
    │           ├─ user approves → proceed
    │           └─ user denies   → skip tool, send denial
    │
    ├─ tools.Registry.Get(name).Execute(ctx, args)
    │     │
    │     └─ returns Result{Content, IsError}
    │
    ├─ hooks.Executor.RunPost(ctx, HookContext{ToolName, Args, Result})
    │
    └─ append tool_result to messages:
          Message{
              Role: "user",
              Content: ContentBlock{
                  Type: "tool_result",
                  ID: toolCallID,
                  Content: result.Content,
                  IsError: result.IsError,
              }
          }
```

### 5. Loop Continuation

```
After all tool results appended:
    │
    ├─ increment turn counter
    │
    ├─ if turn >= maxTurns: break with warning
    │
    ├─ check token budget:
    │     if total tokens > compactionThreshold:
    │         compact older messages (summarize or truncate)
    │
    └─ loop back to step 2 (api.Client.Stream with updated messages)
```

## Streaming Protocol

### Anthropic SSE Format

```
event: message_start
data: {"type":"message_start","message":{"id":"msg_xxx",...}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":15}}

event: message_stop
data: {"type":"message_stop"}
```

### OpenAI SSE Format

```
data: {"choices":[{"delta":{"content":"Hello"},"index":0}]}

data: {"choices":[{"delta":{},"index":0,"finish_reason":"stop"}]}

data: [DONE]
```

Both are normalized to the same `StreamEvent` channel — the engine doesn't know which provider it's using.

## Message Types

```
┌─────────────────────────────────────────────────────────────┐
│ Conversation History                                        │
│                                                             │
│ Message{Role: "user",      Content: "fix the bug"}          │
│ Message{Role: "assistant",  Content: [text, tool_use]}      │
│ Message{Role: "user",      Content: [tool_result]}          │
│ Message{Role: "assistant",  Content: "I've fixed it"}       │
│ Message{Role: "user",      Content: "run the tests"}        │
│ ...                                                         │
└─────────────────────────────────────────────────────────────┘
```

### Content Block Types

| Type | Direction | Fields |
|---|---|---|
| `text` | assistant → user | `text` |
| `tool_use` | assistant → engine | `id`, `name`, `input` |
| `tool_result` | engine → API | `tool_use_id`, `content`, `is_error` |

## Compaction Strategy

When token count exceeds the budget:

1. Keep the system prompt (always)
2. Keep the last N turns (configurable, default 10)
3. Summarize older turns into a single `user` message:
   ```
   "Previous conversation summary: The user asked to fix a bug in auth.go.
   The assistant edited the file and ran tests. Tests passed."
   ```
4. Replace old messages with the summary
5. Continue the loop with compacted history

## Permission Check Sequence

```
Tool call arrives
    │
    ├─ Is tool in explicit deny list? → Deny
    │
    ├─ Is tool in explicit allow list? → Allow
    │
    ├─ Check path rules (glob match on file args):
    │     ├─ path matches deny pattern → Deny
    │     └─ path matches allow pattern → Allow
    │
    ├─ Check command deny patterns (for bash tool):
    │     └─ command matches pattern → Deny
    │
    └─ Fall through to mode-based default:
          ├─ default: Ask for write tools, Allow for read-only
          ├─ plan:    Deny all write/shell tools
          └─ auto:    Allow all (except deny list)
```

## Hook Execution Sequence

```
Plugin hooks loaded from:
    ~/.openharness/plugins/*/hooks.json
    ./.openharness/plugins/*/hooks.json

Each hooks.json:
    {
      "hooks": [
        {"event": "PreToolUse",  "command": "check-safety.sh"},
        {"event": "PostToolUse", "command": "log-usage.sh"}
      ]
    }

Pre-tool hooks (in registration order):
    1. Built-in hooks (engine registered)
    2. Plugin hooks (discovery order)
    → First Block wins, others are skipped

Post-tool hooks (all run, order not guaranteed):
    1. Built-in hooks
    2. Plugin hooks
    → Errors logged, not fatal
```
