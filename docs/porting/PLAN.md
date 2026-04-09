# Porting Plan

Concrete plan for porting OpenHarness (Python) to ohgo (Go), organized by phase with dependency ordering.

**Status legend:** `TODO` `IN PROGRESS` `DONE` `SKIPPED`

---

## Phase 1: Foundation

The minimal skeleton that compiles and can make a single LLM API call.

| # | Python Source | Go Target | Key Types to Port | Depends On | Unit Test | Status |
|---|---|---|---|---|---|---|
| 1.1 | `config/paths.py` | `internal/config/paths.go` | `ConfigDir()`, `DataDir()`, `ProjectDir()` | — | `paths_test.go`: verify paths exist, env override | DONE |
| 1.2 | `config/schema.py` | `internal/config/schema.go` | `Settings`, `ProviderProfile`, `PermissionSettings` | 1.1 | `schema_test.go`: JSON round-trip, validation tags | DONE |
| 1.3 | `config/settings.py` | `internal/config/settings.go` | `Load()`, `Merge()`, layered config resolution | 1.2 | `settings_test.go`: merge precedence, env override, missing file | DONE |
| 1.4 | `platforms.py` | `internal/config/platform.go` | `OS`, `Arch`, `Shell`, `WorkingDir()` | — | `platform_test.go`: detect current OS, shell, git repo | DONE |
| 1.5 | `engine/messages.py` | `internal/engine/message.go` | `Message`, `ContentBlock`, `TextBlock`, `ToolUseBlock`, `ToolResultBlock` | — | `message_test.go`: JSON marshal/unmarshal all block types | DONE |
| 1.6 | `engine/stream_events.py` | `internal/engine/events.go` | `AssistantTextDelta`, `ToolExecutionStarted`, `ToolExecutionCompleted`, `AssistantTurnComplete` | 1.5 | `events_test.go`: event type discrimination, payload access | DONE |
| 1.7 | `api/errors.py` | `internal/api/errors.go` | `APIError`, `RateLimitError`, `AuthError`, sentinel errors | — | `errors_test.go`: errors.Is, errors.As matching, retryable check | DONE |
| 1.8 | `api/client.py` | `internal/api/anthropic.go` | `AnthropicClient`, `StreamMessage()`, retry with backoff | 1.5, 1.6, 1.7 | `anthropic_test.go`: mock SSE server, stream parsing, retry on 429, context cancel | DONE |
| 1.9 | `api/usage.py` | `internal/api/usage.go` | `UsageSnapshot`, `TokenCounts`, cost tracking | 1.6 | `usage_test.go`: token aggregation, cost calculation | DONE |
| 1.10 | `tools/base.py` | `internal/tools/tool.go` | `Tool` interface, `Result`, `Registry` | — | `tool_test.go`: register, get, list, duplicate panic, concurrent access | DONE |
| 1.11 | `engine/query_engine.py` | `internal/engine/engine.go` | `QueryEngine`, `QueryContext`, `Query()` agent loop | 1.5, 1.6, 1.8, 1.10 | `engine_test.go`: mock API client, verify tool_use loop, max turns, context cancel | DONE |
| 1.12 | `engine/cost_tracker.go` | `internal/engine/cost.go` | `CostTracker`, turn counting, usage aggregation | 1.6, 1.9 | `cost_test.go`: accumulate usage, track turns | DONE |
| 1.13 | `cli.py` | `cmd/og/main.go` | cobra root command, flag parsing, REPL entry | 1.3, 1.11 | see manual test below | DONE |

### Phase 1 Manual Test

```bash
# Build
go build ./cmd/og

# 1. Help output
./og --help
# Expect: shows --model, --prompt, --permission flags

# 2. Config loading
./og --prompt "hello" 2>&1 | head -5
# Expect: loads config, connects to API (may fail without key — that's ok)

# 3. Streaming call (requires ANTHROPIC_API_KEY)
export ANTHROPIC_API_KEY=sk-...
./og --prompt "Say hello in one word" --model claude-sonnet-4-6-20250514
# Expect: streaming text output, single word response

# 4. Context cancel
./og --prompt "Count to 1000 slowly" &
sleep 1 && kill -INT $!
# Expect: clean shutdown, no goroutine leak
```

---

## Phase 2: Safety & Prompts

Permission checks, hooks, and system prompt assembly — everything that wraps tool execution.

| # | Python Source | Go Target | Key Types to Port | Depends On | Status |
|---|---|---|---|---|---|
| 2.1 | `permissions/modes.py` | `internal/permissions/modes.go` | `Mode` enum, mode behavior matrix | — | DONE |
| 2.2 | `permissions/checker.py` | `internal/permissions/checker.go` | `Checker` interface, `DefaultChecker`, path rules, command deny patterns | 2.1 | DONE |
| 2.3 | `hooks/types.py` + `hooks/schemas.py` | `internal/hooks/types.go` | `HookType` (Command/HTTP/Prompt/Agent), `HookDefinition`, pattern matching | — | DONE |
| 2.4 | `hooks/loader.py` | `internal/hooks/loader.go` | `LoadFromDir()`, parse `hooks.json` manifests | 2.3 | DONE |
| 2.5 | `hooks/executor.py` | `internal/hooks/executor.go` | `Executor`, `RunPre()`, `RunPost()`, aggregation | 2.3, 2.4 | DONE |
| 2.6 | `hooks/hot_reload.py` | `internal/hooks/reload.go` | `WatchAndReload()`, fsnotify integration | 2.4 | DONE |
| 2.7 | `prompts/environment.py` | `internal/prompts/environment.go` | `EnvironmentInfo{OS, Shell, Cwd, GitStatus, Date}` | 1.4 | DONE |
| 2.8 | `prompts/system_prompt.py` | `internal/prompts/system.go` | `BuildSystemPrompt()`, base prompt template | 2.7 | DONE |
| 2.9 | `prompts/claudemd.py` | `internal/prompts/claudemd.go` | `DiscoverCLAUDEmd()`, walk up to root, merge | — | DONE |
| 2.10 | `prompts/context.py` | `internal/prompts/context.go` | `BuildContextPrompt()`, inject environment + CLAUDE.md | 2.8, 2.9 | DONE |
| 2.11 | `services/compact/` | `internal/engine/compact.go` | `Microcompact()`, `FullCompact()`, token budget check | 1.5, 1.11 | DONE |

**Phase 1 milestone:** `og` starts, loads config, sends one prompt to Claude API, prints streaming response.

---

## Phase 2: Safety & Prompts

Permission checks, hooks, and system prompt assembly — everything that wraps tool execution.

| # | Python Source | Go Target | Key Types to Port | Depends On | Unit Test | Status |
|---|---|---|---|---|---|---|
| 2.1 | `permissions/modes.py` | `internal/permissions/modes.go` | `Mode` enum, mode behavior matrix | — | `modes_test.go`: all 3 modes, read-only vs write tool classification | DONE |
| 2.2 | `permissions/checker.py` | `internal/permissions/checker.go` | `Checker` interface, `DefaultChecker`, path rules, command deny patterns | 2.1 | `checker_test.go`: allow/deny/ask per mode, glob path rules, command deny patterns, explicit tool lists | DONE |
| 2.3 | `hooks/types.py` + `hooks/schemas.py` | `internal/hooks/types.go` | `HookType` (Command/HTTP/Prompt/Agent), `HookDefinition`, pattern matching | — | `types_test.go`: fnmatch matching, JSON unmarshal hook defs | DONE |
| 2.4 | `hooks/loader.py` | `internal/hooks/loader.go` | `LoadFromDir()`, parse `hooks.json` manifests | 2.3 | `loader_test.go`: load valid manifest, missing file, malformed JSON, nested dirs | DONE |
| 2.5 | `hooks/executor.py` | `internal/hooks/executor.go` | `Executor`, `RunPre()`, `RunPost()`, aggregation | 2.3, 2.4 | `executor_test.go`: pre-hook block, pre-hook modify args, post-hook error logging, multiple hooks ordering | DONE |
| 2.6 | `hooks/hot_reload.py` | `internal/hooks/reload.go` | `WatchAndReload()`, fsnotify integration | 2.4 | `reload_test.go`: write hooks.json, verify reload fires | DONE |
| 2.7 | `prompts/environment.py` | `internal/prompts/environment.go` | `EnvironmentInfo{OS, Shell, Cwd, GitStatus, Date}` | 1.4 | `environment_test.go`: detect OS/shell, git status in repo vs non-repo | DONE |
| 2.8 | `prompts/system_prompt.py` | `internal/prompts/system.go` | `BuildSystemPrompt()`, base prompt template | 2.7 | `system_test.go`: prompt contains environment info, custom prompt override | DONE |
| 2.9 | `prompts/claudemd.py` | `internal/prompts/claudemd.go` | `DiscoverCLAUDEmd()`, walk up to root, merge | — | `claudemd_test.go`: discover in nested dir, merge user+project, missing files | DONE |
| 2.10 | `prompts/context.py` | `internal/prompts/context.go` | `BuildContextPrompt()`, inject environment + CLAUDE.md | 2.8, 2.9 | `context_test.go`: full prompt assembly, empty CLAUDE.md | DONE |
| 2.11 | `services/compact/` | `internal/engine/compact.go` | `Microcompact()`, `FullCompact()`, token budget check | 1.5, 1.11 | `compact_test.go`: microcompact removes old tool results, budget exceeded detection | DONE |

### Phase 2 Manual Test

```bash
# 1. Permission prompt
./og --prompt "Create a file called /tmp/test.txt with hello"
# Expect: permission prompt appears, approve → file created, deny → blocked message

# 2. Plan mode blocks writes
./og --prompt "Create a file called /tmp/test.txt" --permission plan
# Expect: tool denied, no file created

# 3. Auto mode allows writes
./og --prompt "Create a file called /tmp/test.txt with hello" --permission auto
# Expect: file created without prompt

# 4. System prompt includes CLAUDE.md
echo "Always respond in French" > CLAUDE.md
./og --prompt "Say hello"
# Expect: response in French
rm CLAUDE.md

# 5. Hook blocking
mkdir -p .openharness/plugins/test
cat > .openharness/plugins/test/hooks.json << 'EOF'
{"hooks":[{"event":"PreToolUse","pattern":"bash","command":"echo BLOCKED >&2 && exit 1"}]}
EOF
./og --prompt "Run ls"
# Expect: bash tool blocked by hook
rm -rf .openharness
```

---

## Phase 3: Essential Tools (first 10)

The tools needed for a usable coding agent. Each tool is a self-contained package under `internal/tools/<name>/`.

| # | Python Source | Go Target | Description | Depends On | Unit Test | Status |
|---|---|---|---|---|---|---|
| 3.1 | `tools/file_read_tool.py` | `internal/tools/read/read.go` | Read file contents with line numbers | Phase 2 | `read_test.go`: read full file, line range, offset/limit, missing file, binary file, empty file | DONE |
| 3.2 | `tools/file_write_tool.py` | `internal/tools/write/write.go` | Write/create files | Phase 2 | `write_test.go`: create new, overwrite existing, create parent dirs, permission error | DONE |
| 3.3 | `tools/file_edit_tool.py` | `internal/tools/edit/edit.go` | Exact string replacement in files | Phase 2 | `edit_test.go`: exact match, no match error, multiple matches error, empty old_string, multiline replacement | DONE |
| 3.4 | `tools/bash_tool.py` | `internal/tools/bash/bash.go` | Shell command execution with timeout | Phase 2 | `bash_test.go`: echo command, exit code, stderr, timeout, context cancel, working dir | DONE |
| 3.5 | `tools/glob_tool.py` | `internal/tools/glob/glob.go` | File pattern matching | Phase 2 | `glob_test.go`: **/*.go, single *, no matches, nested dirs, symlink | DONE |
| 3.6 | `tools/grep_tool.py` | `internal/tools/grep/grep.go` | Content search (regex) | Phase 2 | `grep_test.go`: pattern match, case insensitive, context lines, file filter, no matches | DONE |
| 3.7 | `tools/web_fetch_tool.py` | `internal/tools/webfetch/fetch.go` | HTTP GET, HTML→text conversion | stdlib net/http | `fetch_test.go`: mock HTTP server, HTML stripping, JSON response, timeout, 404 | DONE |
| 3.8 | `tools/web_search_tool.py` | `internal/tools/websearch/search.go` | Web search via DuckDuckGo | stdlib net/http | `search_test.go`: mock search API, parse results, empty results, API error | DONE |
| 3.9 | `tools/register.go` | `internal/tools/builtin/builtin.go` | `RegisterAll()` — register all tools | 3.1–3.8 | `builtin_test.go`: all tools registered, no duplicates, each has valid schema | DONE |
| 3.10 | `tools/lsp_tool.py` | `internal/tools/lsp/lsp.go` | LSP operations (stub — validates inputs) | 3.1 | `lsp_test.go`: valid operations, invalid operation, missing required fields | DONE |

### Phase 3 Manual Test

```bash
# 1. Read a file
./og --prompt "Read the file go.mod and tell me the module name"
# Expect: reads go.mod, reports "github.com/shtdu/ohgo"

# 2. Write a file
./og --prompt "Create a file /tmp/og-test.txt with content 'hello from og'"
# Expect: file created, verify: cat /tmp/og-test.txt

# 3. Edit a file
./og --prompt "Edit /tmp/og-test.txt and change 'hello' to 'goodbye'"
# Expect: file updated, verify: cat /tmp/og-test.txt shows 'goodbye from og'

# 4. Run a command
./og --prompt "Run 'ls go.mod' and tell me if it exists"
# Expect: runs ls, reports file exists

# 5. Search files
./og --prompt "Find all .go files in this project using glob"
# Expect: lists all .go files

# 6. Search content
./og --prompt "Search for 'func main' in all Go files"
# Expect: finds main functions in cmd/og and cmd/ogmo

# 7. Multi-tool workflow
./og --prompt "Read go.mod, then create /tmp/og-module.txt with just the module name"
# Expect: reads go.mod, extracts module name, writes file
```

**Phase 3 milestone:** `og` can read, write, edit files, run commands, and search code. Usable as a basic coding agent.

---

## Phase 4: Extended Tools

Remaining tools for full feature parity.

| # | Python Source | Go Target | Description | Depends On | Unit Test | Status |
|---|---|---|---|---|---|---|
| 4.1 | `tools/agent_tool.py` | `internal/tools/agent/agent.go` | Spawn subagent | 3.9, coordinator | `agent_test.go`: mock engine, verify prompt passed, result returned | TODO |
| 4.2 | `tools/ask_user_question_tool.py` | `internal/tools/ask/ask.go` | Interactive user question | UI | `ask_test.go`: mock UI, verify question rendered, answer returned | TODO |
| 4.3 | `tools/brief_tool.py` | `internal/tools/brief/brief.go` | Brief mode toggle | — | `brief_test.go`: toggle state | DONE |
| 4.4 | `tools/config_tool.py` | `internal/tools/config/config.go` | Read/write config from tool | 1.3 | `config_test.go`: read key, write key, missing key | DONE |
| 4.5 | `tools/cron_create_tool.py` | `internal/tools/cron/create.go` | Create cron job | robfig/cron | `create_test.go`: valid cron expr, invalid expr, duplicate | DONE |
| 4.6 | `tools/cron_delete_tool.py` | `internal/tools/cron/delete.go` | Delete cron job | robfig/cron | `delete_test.go`: delete existing, delete missing | DONE |
| 4.7 | `tools/cron_list_tool.py` | `internal/tools/cron/list.go` | List cron jobs | robfig/cron | `list_test.go`: empty list, multiple jobs | DONE |
| 4.8 | `tools/cron_toggle_tool.py` | `internal/tools/cron/toggle.go` | Enable/disable cron job | robfig/cron | `toggle_test.go`: toggle on→off, off→on | DONE |
| 4.9 | `tools/enter_plan_mode_tool.py` | `internal/tools/plan/enter.go` | Enter plan mode | 2.1 | `enter_test.go`: mode changes to plan | DONE |
| 4.10 | `tools/exit_plan_mode_tool.py` | `internal/tools/plan/exit.go` | Exit plan mode | 2.1 | `exit_test.go`: mode changes back to default | DONE |
| 4.11 | `tools/enter_worktree_tool.py` | `internal/tools/worktree/enter.go` | Create git worktree | — | `enter_test.go`: mock git, verify worktree created | DONE |
| 4.12 | `tools/exit_worktree_tool.py` | `internal/tools/worktree/exit.go` | Leave git worktree | — | `exit_test.go`: mock git, verify cleanup | DONE |
| 4.13 | `tools/notebook_edit_tool.py` | `internal/tools/notebook/edit.go` | Jupyter notebook cell editing | — | `edit_test.go`: read .ipynb, replace cell, insert cell, delete cell | DONE |
| 4.14 | `tools/sleep_tool.py` | `internal/tools/sleep/sleep.go` | Delay execution | — | `sleep_test.go`: short sleep completes, context cancel interrupts | DONE |
| 4.15 | `tools/send_message_tool.py` | `internal/tools/message/message.go` | Send message to user | — | `message_test.go`: message captured | TODO |
| 4.16 | `tools/skill_tool.py` | `internal/tools/skill/skill.go` | Invoke a skill | skills package | `skill_test.go`: mock loader, verify skill loaded | DONE |
| 4.17 | `tools/remote_trigger_tool.py` | `internal/tools/remote/trigger.go` | Trigger remote action | — | `trigger_test.go`: mock HTTP, verify request | DONE |
| 4.18 | `tools/todo_write_tool.py` | `internal/tools/todo/todo.go` | Write todo list | — | `todo_test.go`: write todos, clear todos | DONE |
| 4.19 | `tools/tool_search_tool.py` | `internal/tools/search/search.go` | Search available tools | 3.9 | `search_test.go`: find by name, find by description, no match | DONE |

### Task Tools

| # | Python Source | Go Target | Description | Depends On | Unit Test | Status |
|---|---|---|---|---|---|---|
| 4.20 | `tools/task_create_tool.py` | `internal/tools/task/create.go` | Create background task | tasks package | `create_test.go`: valid create, missing command | DONE |
| 4.21 | `tools/task_get_tool.py` | `internal/tools/task/get.go` | Get task by ID | tasks package | `get_test.go`: existing task, missing ID | DONE |
| 4.22 | `tools/task_list_tool.py` | `internal/tools/task/list.go` | List all tasks | tasks package | `list_test.go`: empty list, multiple tasks | DONE |
| 4.23 | `tools/task_output_tool.go` | `internal/tools/task/output.go` | Read task output | tasks package | `output_test.go`: running task, completed task, missing task | DONE |
| 4.24 | `tools/task_stop_tool.py` | `internal/tools/task/stop.go` | Stop running task | tasks package | `stop_test.go`: stop running, stop already stopped | DONE |
| 4.25 | `tools/task_update_tool.py` | `internal/tools/task/update.go` | Update task status | tasks package | `update_test.go`: status transitions | DONE |

### Team Tools

| # | Python Source | Go Target | Description | Depends On | Unit Test | Status |
|---|---|---|---|---|---|---|
| 4.26 | `tools/team_create_tool.py` | `internal/tools/team/create.go` | Create agent team | coordinator | `create_test.go`: valid team, duplicate name | TODO |
| 4.27 | `tools/team_delete_tool.py` | `internal/tools/team/delete.go` | Delete agent team | coordinator | `delete_test.go`: delete existing, delete missing | TODO |

### MCP Tools

| # | Python Source | Go Target | Description | Depends On | Unit Test | Status |
|---|---|---|---|---|---|---|
| 4.28 | `tools/mcp_tool.py` | `internal/tools/mcp/call.go` | Call MCP server tool | mcp package | `call_test.go`: mock MCP server, valid call, server error | TODO |
| 4.29 | `tools/list_mcp_resources_tool.py` | `internal/tools/mcp/list.go` | List MCP resources | mcp package | `list_test.go`: mock server with resources, empty server | TODO |
| 4.30 | `tools/read_mcp_resource_tool.py` | `internal/tools/mcp/read.go` | Read MCP resource | mcp package | `read_test.go`: mock server, read valid resource | TODO |
| 4.31 | `tools/mcp_auth_tool.py` | `internal/tools/mcp/auth.go` | MCP authentication | mcp package | `auth_test.go`: mock auth flow | TODO |

### Phase 4 Manual Test

```bash
# 1. Plan mode tools
./og --prompt "Enter plan mode, then read go.mod, then exit plan mode"
# Expect: enters plan, reads file, exits plan

# 2. Todo tool
./og --prompt "Create a todo list with 3 items for testing the og tool"
# Expect: todo list created

# 3. Ask user tool
./og --prompt "Ask me what my favorite color is"
# Expect: interactive prompt appears, answer fed back to model

# 4. Cron tools
./og --prompt "Create a cron job that runs 'date' every hour, list all cron jobs"
# Expect: cron created and listed

# 5. MCP tool (requires MCP server configured)
# Skip if no MCP server available

# 6. Background task
./og --prompt "Start a background task that runs 'sleep 5 && echo done', then list tasks"
# Expect: task started, appears in list
```

**Phase 4 milestone:** All 43 tools ported. Full tool-use capability.

---

## Phase 5: UI & Commands

Terminal rendering and slash commands.

| # | Python Source | Go Target | Description | Depends On | Unit Test | Status |
|---|---|---|---|---|---|---|
| 5.1 | `ui/output.py` | `internal/ui/output.go` | Markdown rendering (glamour), styled output (lipgloss) | — | `output_test.go`: render markdown, render code block, render error, empty input | DONE |
| 5.2 | `ui/input.py` | `internal/ui/input.go` | Interactive input with history (bubbletea/huh) | — | `input_test.go`: mock stdin, history navigation, empty input, multiline | DONE |
| 5.3 | `ui/permission_dialog.py` | `internal/ui/permission.go` | Permission approval dialog | 5.2 | `permission_test.go`: mock approve/deny/skip, timeout defaults | TODO |
| 5.4 | `keybindings/` | `internal/ui/keybind.go` | Keybinding parser and resolver | — | `keybind_test.go`: parse binding config, resolve key sequence, default fallback | DONE |
| 5.5 | `output_styles/` | `internal/ui/styles.go` | Output style presets | 5.1 | `styles_test.go`: all preset names valid, each produces non-empty output | DONE |

### Slash Commands (54 total)

Commands are tested via a command registry test harness — mock the engine/UI, verify output.

| # | Command | Go Target | Description | Depends On | Unit Test | Status |
|---|---|---|---|---|---|---|
| 5.6 | `help` | `internal/commands/help.go` | Show available commands | 5.5 | output contains all registered command names | DONE |
| 5.7 | `exit` | `internal/commands/exit.go` | Exit the REPL | — | sets exit flag | DONE |
| 5.8 | `clear` | `internal/commands/clear.go` | Clear conversation history | 1.11 | calls engine.Clear(), verify empty history | DONE |
| 5.9 | `version` | `internal/commands/version.go` | Show version | — | output matches build version | DONE |
| 5.10 | `status` | `internal/commands/status.go` | Session status | 1.11 | output has model, turns, tokens | DONE |
| 5.11 | `context` | `internal/commands/context.go` | Show system prompt | 2.10 | output matches assembled prompt | DONE |
| 5.12 | `summary` | `internal/commands/summary.go` | Summarize conversation | 1.11 | returns summary text | DONE |
| 5.13 | `compact` | `internal/commands/compact.go` | Force compaction | 2.11 | calls compact, verify shorter history | DONE |
| 5.14 | `cost` | `internal/commands/cost.go` | Token usage and cost | 1.12 | output has token counts and estimated cost | DONE |
| 5.15 | `usage` | `internal/commands/usage.go` | Usage stats | 1.9 | output has usage snapshot | DONE |
| 5.16 | `stats` | `internal/commands/stats.go` | Session statistics | 1.11 | output has turn count, duration | DONE |
| 5.17 | `memory` | `internal/commands/memory.go` | Inspect project memory | memory pkg | list memory entries, show entry content | DONE |
| 5.18 | `hooks` | `internal/commands/hooks.go` | Show configured hooks | 2.5 | lists registered hooks | DONE |
| 5.19 | `resume` | `internal/commands/resume.go` | Restore saved session | session storage | load from temp dir, verify history restored | DONE |
| 5.20 | `session` | `internal/commands/session.go` | Inspect session storage | session storage | show session path and size | DONE |
| 5.21 | `export` | `internal/commands/export.go` | Export transcript | — | writes JSON to temp file, verify structure | DONE |
| 5.22 | `share` | `internal/commands/share.go` | Shareable transcript | — | creates file, verify content | DONE |
| 5.23 | `copy` | `internal/commands/copy.go` | Copy to clipboard | atotto/clipboard | mock clipboard, verify content written | DONE |
| 5.24 | `tag` | `internal/commands/tag.go` | Named session snapshot | session storage | create tag, list tags | DONE |
| 5.25 | `rewind` | `internal/commands/rewind.go` | Remove last turn(s) | 1.11 | rewind 1 turn, rewind N turns, rewind past start | DONE |
| 5.26 | `files` | `internal/commands/files.go` | List workspace files | — | list files in test dir | DONE |
| 5.27 | `init` | `internal/commands/init.go` | Initialize project files | — | creates .openharness dir, verify files | DONE |
| 5.28 | `bridge` | `internal/commands/bridge.go` | Inspect bridge helpers | bridge pkg | show bridge status | DONE |
| 5.29 | `login` / `logout` | `internal/commands/auth.go` | Auth status, store/clear API key | auth pkg | mock key store, verify save/clear | DONE |
| 5.30 | `feedback` | `internal/commands/feedback.go` | Save feedback | — | writes to file | DONE |
| 5.31 | `onboarding` | `internal/commands/onboarding.go` | Quickstart guide | — | output non-empty | DONE |
| 5.32 | `skills` | `internal/commands/skills.go` | List/show skills | skills pkg | mock loader, list skills | DONE |
| 5.33 | `config` | `internal/commands/config.go` | Show/update config | 1.3 | show key, set key, invalid key | DONE |
| 5.34 | `mcp` | `internal/commands/mcp.go` | MCP status | mcp pkg | show connected servers | DONE |
| 5.35 | `plugin` | `internal/commands/plugin.go` | Manage plugins | plugins pkg | list, install, remove | DONE |
| 5.36 | `reload-plugins` | `internal/commands/reload.go` | Reload plugins | plugins pkg | verify re-scan | DONE |
| 5.37 | `permissions` | `internal/commands/perms.go` | Show/update permissions | 2.2 | show mode, set mode | DONE |
| 5.38 | `plan` | `internal/commands/plan.go` | Toggle plan mode | 2.1 | toggle on/off | DONE |
| 5.39 | `fast` | `internal/commands/fast.go` | Fast mode toggle | — | toggle state | DONE |
| 5.40 | `effort` | `internal/commands/effort.go` | Reasoning effort | — | set valid effort | DONE |
| 5.41 | `passes` | `internal/commands/passes.go` | Reasoning passes | — | set valid count | DONE |
| 5.42 | `turns` | `internal/commands/turns.go` | Max turn count | 1.11 | set valid count, reject 0 | DONE |
| 5.43 | `continue` | `internal/commands/cont.go` | Continue tool loop | 1.11 | resumes engine | DONE |
| 5.44 | `provider` | `internal/commands/provider.go` | Switch provider profiles | 1.3 | list, switch profile | DONE |
| 5.45 | `model` | `internal/commands/model.go` | Switch model | 1.3 | list, switch model | DONE |
| 5.46 | `theme` | `internal/commands/theme.go` | TUI themes | 5.1 | list themes, set valid theme | DONE |
| 5.47 | `output-style` | `internal/commands/style.go` | Output style | 5.5 | list styles, set valid style | DONE |
| 5.48 | `keybindings` | `internal/commands/keybind.go` | Show keybindings | 5.4 | output bindings | DONE |
| 5.49 | `vim` | `internal/commands/vim.go` | Vim mode toggle | 5.2 | toggle state | DONE |
| 5.50 | `voice` | `internal/commands/voice.go` | Voice mode toggle | — | toggle state | DONE |
| 5.51 | `doctor` | `internal/commands/doctor.go` | Environment diagnostics | 1.4 | output has OS, shell, go version | DONE |
| 5.52 | `diff` | `internal/commands/diff.go` | Git diff | — | output matches `git diff` | DONE |
| 5.53 | `branch` | `internal/commands/branch.go` | Git branch info | — | output has current branch | DONE |
| 5.54 | `commit` | `internal/commands/commit.go` | Git commit workflow | — | mock git, verify commit message | DONE |
| 5.55 | `issue` | `internal/commands/issue.go` | Issue context | — | output issue list | DONE |
| 5.56 | `pr_comments` | `internal/commands/pr.go` | PR comments context | — | mock gh, output comments | DONE |
| 5.57 | `privacy-settings` | `internal/commands/privacy.go` | Privacy settings | — | show, toggle | DONE |
| 5.58 | `release-notes` | `internal/commands/releases.go` | Release notes | — | output non-empty | DONE |
| 5.59 | `upgrade` | `internal/commands/upgrade.go` | Upgrade instructions | — | output non-empty | DONE |
| 5.60 | `agents` | `internal/commands/agents.go` | List agent tasks | coordinator | output agent list | DONE |
| 5.61 | `tasks` | `internal/commands/tasks.go` | Manage background tasks | tasks pkg | output task list | DONE |

### Phase 5 Manual Test

```bash
# Start interactive REPL
./og

# In the REPL, test these slash commands:
/help                       # Expect: list of all commands
/version                    # Expect: version number
/status                     # Expect: model, turns, tokens
/config                     # Expect: current config
/model                      # Expect: current model name
/permissions                # Expect: current permission mode
/plan                       # Expect: enters plan mode
/plan                       # Expect: exits plan mode
/doctor                     # Expect: environment diagnostics
/diff                       # Expect: git diff output
/branch                     # Expect: current branch
/cost                       # Expect: token usage
/clear                      # Expect: conversation cleared
/exit                       # Expect: clean exit

# Test interactive prompt rendering
# Type a prompt, verify markdown renders correctly (bold, code, links)
# Verify streaming text appears incrementally
# Verify tool execution shows progress spinner
```

**Phase 5 milestone:** Full interactive REPL with all slash commands and rich terminal output.

---

## Phase 6: Subsystems

Self-contained subsystems that the engine and tools depend on.

| # | Python Source | Go Target | Key Types to Port | Depends On | Unit Test | Status |
|---|---|---|---|---|---|---|
| 6.1 | `skills/` | `internal/skills/skills.go` + `parse.go` + `registry.go` | `Loader`, `Load()`, `LoadAll()`, YAML frontmatter parsing, `Registry` | 1.3 | `skills_test.go`, `parse_test.go`, `registry_test.go` | DONE |
| 6.2 | `skills/` | `internal/skills/skills.go` | Skill discovery from multiple directories | 6.1 | `skills_test.go`: discover in multiple dirs, dedup, empty dirs | DONE |
| 6.3 | `memory/paths.py` | `internal/memory/paths.go` | `ProjectDir()`, `Entrypoint()` | 1.1 | `paths_test.go`: verify paths constructed correctly | DONE |
| 6.4 | `memory/types.py` + `scan.py` | `internal/memory/scan.go` + `types.go` | `Header`, `Scan()` | 6.3 | `scan_test.go`: scan temp dir, parse headers, empty dir | DONE |
| 6.5 | `memory/manager.py` + `memdir.py` | `internal/memory/memory.go` | `Store.Add()`, `Store.Remove()`, `Store.List()`, `Store.LoadPrompt()`, `MEMORY.md` index | 6.3, 6.4 | `memory_test.go`: add+list round-trip, remove updates index, load prompt | DONE |
| 6.6 | `memory/search.py` | `internal/memory/search.go` | `Find()`, relevance scoring with ASCII + Han tokenization | 6.4 | `search_test.go`: keyword match, no match, Han chars | DONE |
| 6.7 | `mcp/config.py` | `internal/mcp/config.go` | `McpServerConfig`, load from settings | 1.3 | `config_test.go`: parse valid config, missing server, invalid transport | TODO |
| 6.8 | `mcp/types.py` | `internal/mcp/types.go` | `McpTool`, `McpResource`, protocol types | — | `types_test.go`: JSON round-trip all types | TODO |
| 6.9 | `mcp/client.py` | `internal/mcp/client.go` | `Client`, `Connect()`, `CallTool()`, `ListTools()` | 6.7, 6.8, mcp-go | `client_test.go`: mock stdio MCP server, connect, list tools, call tool, error handling | TODO |
| 6.10 | `plugins/types.py` + `schemas.py` | `internal/plugins/types.go` | `Manifest`, `LoadedPlugin` | — | `types_test.go`: parse valid manifest, missing fields, JSON round-trip | DONE |
| 6.11 | `plugins/loader.py` | `internal/plugins/loader.go` | `Discover()`, directory scanning | 6.10 | `loader_test.go`: scan temp plugin dir, nested plugins, invalid plugin.json | DONE |
| 6.12 | `plugins/installer.py` | `internal/plugins/installer.go` | `Install()`, `Uninstall()` | 6.10 | `installer_test.go`: install from source, uninstall by name, missing plugin | DONE |
| 6.13 | `coordinator/agent_definitions.py` | `internal/coordinator/defs.go` | `AgentDefinition`, YAML loading | 6.1 | `defs_test.go`: parse valid YAML, tool filtering, model override | TODO |
| 6.14 | `coordinator/coordinator_mode.py` | `internal/coordinator/mode.go` | `CoordinatorMode`, agent lifecycle | 6.13 | `mode_test.go`: spawn agent, verify isolation, cleanup | TODO |
| 6.15 | `coordinator/registry.py` | `internal/coordinator/registry.go` | Agent registry, spawning | 6.14 | `registry_test.go`: register, lookup, list, concurrent access | TODO |
| 6.16 | `tasks/` | `internal/tasks/manager.go` + `types.go` | `Manager`, `CreateShell()`, `Stop()`, `ReadOutput()`, subprocess lifecycle | — | `manager_test.go`: start task, get output, stop task, list tasks, context cancel | DONE |
| 6.17 | `sandbox/adapter.py` | `internal/sandbox/sandbox.go` | `Availability`, `CheckAvailability()`, `WrapCommand()` | — | `sandbox_test.go`: availability check, wrap command, config generation | DONE |

### Phase 6 Manual Test

```bash
# 1. Skills
mkdir -p ~/.openharness/skills
cat > ~/.openharness/skills/test.md << 'EOF'
---
name: test-skill
description: A test skill
---
Do something useful.
EOF
./og --prompt "Use the test-skill"
# Expect: skill loaded and used

# 2. Memory
./og --prompt "Remember that I prefer Go over Python"
# Expect: memory saved
./og --prompt "What language do I prefer?"
# Expect: recalls Go preference from memory

# 3. MCP (requires MCP server)
# Configure an MCP server in settings.json, then:
./og --prompt "List MCP tools"
# Expect: shows tools from MCP server

# 4. Plugins
mkdir -p .openharness/plugins/test-plugin
cat > .openharness/plugins/test-plugin/plugin.json << 'EOF'
{"name":"test","version":"1.0","description":"Test plugin"}
EOF
./og
# In REPL: /plugin
# Expect: lists test-plugin
```

**Phase 6 milestone:** All subsystems operational. Skills, memory, MCP, plugins, coordinator, tasks.

---

## Phase 7: API Providers

Additional LLM provider clients and bridges.

| # | Python Source | Go Target | Description | Depends On | Unit Test | Status |
|---|---|---|---|---|---|---|
| 7.1 | `api/openai_client.py` | `internal/api/openai.go` | OpenAI-compatible streaming client | 1.5, 1.6 | `openai_test.go`: mock SSE server (OpenAI format), stream parsing, tool_use, retry, context cancel | TODO |
| 7.2 | `api/provider.py` + `registry.py` | `internal/api/registry.go` | Provider factory, profile → client mapping | 7.1, 1.3 | `registry_test.go`: create client by profile type, unknown profile, missing API key | TODO |
| 7.3 | `api/copilot_client.py` | `internal/api/copilot.go` | GitHub Copilot client | 7.1, 7.4 | `copilot_test.go`: mock Copilot token exchange, streaming | TODO |
| 7.4 | `api/copilot_auth.py` | `internal/auth/copilot.go` | Copilot OAuth device flow | 1.7 | `copilot_test.go`: mock OAuth endpoints, device code flow, token refresh | TODO |
| 7.5 | `auth/manager.py` + `storage.py` | `internal/auth/manager.go` | Auth manager, key storage | — | `manager_test.go`: store key, load key, delete key, encrypted storage | TODO |
| 7.6 | `auth/flows.py` + `external.py` | `internal/auth/flows.go` | Auth flow orchestration | 7.5 | `flows_test.go`: mock provider auth, flow selection by provider type | TODO |
| 7.7 | `api/codex_client.py` | `internal/api/codex.go` | Codex CLI bridge client | 7.1 | `codex_test.go`: mock Codex CLI binary, streaming, error handling | TODO |
| 7.8 | `bridge/types.py` + `manager.py` | `internal/bridge/manager.go` | Bridge manager, lifecycle | 7.7, 7.3 | `manager_test.go`: register bridge, connect all, close all, bridge not found | TODO |
| 7.9 | `bridge/session_runner.py` | `internal/bridge/session.go` | Bridge session execution | 7.8 | `session_test.go`: run session, capture output, error propagation | TODO |
| 7.10 | `bridge/work_secret.py` | `internal/bridge/secret.go` | Work secret management | 7.8 | `secret_test.go`: generate, validate, rotate | TODO |

### Phase 7 Manual Test

```bash
# 1. OpenAI-compatible provider
./og --provider openai-compatible --model gpt-4 --prompt "Say hello"
# Expect: works with OpenAI API (requires OPENAI_API_KEY)

# 2. Switch provider in REPL
./og
# /provider
# Expect: lists available providers
# /provider openai-compatible
# Expect: switches provider

# 3. Auth login
./og
# /login
# Expect: shows auth status, prompts for API key if missing
```

**Phase 7 milestone:** All provider types work — Anthropic API, OpenAI-compatible, Copilot, Codex bridge, Claude subscription bridge.

---

## Phase 8: Channels (ogmo)

IM channel integrations for the ogmo personal agent.

| # | Python Source | Go Target | Description | Depends On | Unit Test | Status |
|---|---|---|---|---|---|---|
| 8.1 | `channels/adapter.py` + `base.py` | `internal/channels/adapter.go` | Base channel interface, adapter pattern | — | `adapter_test.go`: interface compliance, message conversion | TODO |
| 8.2 | `channels/bus/events.py` + `queue.py` | `internal/channels/bus.go` | Event bus, message queue | 8.1 | `bus_test.go`: publish/subscribe, queue ordering, backpressure | TODO |
| 8.3 | `channels/impl/telegram.py` | `internal/channels/telegram/telegram.go` | Telegram bot integration | 8.2, resty | `telegram_test.go`: mock Telegram API, send/receive, webhook | TODO |
| 8.4 | `channels/impl/slack.py` | `internal/channels/slack/slack.go` | Slack bot integration | 8.2, gorilla/websocket | `slack_test.go`: mock Slack RTM, message handling | TODO |
| 8.5 | `channels/impl/discord.py` | `internal/channels/discord/discord.go` | Discord bot integration | 8.2, gorilla/websocket | `discord_test.go`: mock Discord gateway, message handling | TODO |
| 8.6 | `channels/impl/feishu.py` | `internal/channels/feishu/feishu.go` | Feishu/Lark integration | 8.2, resty | `feishu_test.go`: mock Feishu API, event handling | TODO |
| 8.7 | `channels/impl/dingtalk.py` | `internal/channels/dingtalk/dingtalk.go` | DingTalk integration | 8.2, resty | `dingtalk_test.go`: mock DingTalk API, callback | TODO |
| 8.8 | `channels/impl/matrix.py` | `internal/channels/matrix/matrix.go` | Matrix integration | 8.2 | `matrix_test.go`: mock Matrix client, sync loop | TODO |
| 8.9 | `channels/impl/email.py` | `internal/channels/email/email.go` | Email channel | 8.2 | `email_test.go`: mock SMTP, send/receive | TODO |
| 8.10 | `channels/impl/mochat.py` | `internal/channels/mochat/mochat.go` | Mochat integration | 8.2, resty | `mochat_test.go`: mock Mochat API | TODO |
| 8.11 | `channels/impl/qq.py` | `internal/channels/qq/qq.go` | QQ integration | 8.2 | `qq_test.go`: mock QQ API | TODO |
| 8.12 | `channels/impl/whatsapp.py` | `internal/channels/whatsapp/whatsapp.go` | WhatsApp integration | 8.2, gorilla/websocket | `whatsapp_test.go`: mock WhatsApp Web API | TODO |
| 8.13 | `channels/impl/manager.py` | `internal/channels/manager.go` | Channel lifecycle manager | 8.1–8.12 | `manager_test.go`: register channels, start/stop all, error isolation | TODO |
| 8.14 | `ohmo/cli.py` | `cmd/ogmo/main.go` | ogmo CLI with channel selection | 8.13, engine | see manual test | TODO |
| 8.15 | `ohmo/workspace.py` + `prompts.py` | `internal/channels/workspace.go` | Workspace and bootstrap prompts | — | `workspace_test.go`: workspace init, prompt assembly | TODO |
| 8.16 | `ohmo/gateway/` | `internal/channels/gateway.go` | Gateway service (HTTP router) | 8.13 | `gateway_test.go`: mock HTTP requests, routing, auth | TODO |

### Phase 8 Manual Test

```bash
# 1. ogmo help
./ogmo --help
# Expect: shows --channel flag, supported channels listed

# 2. Telegram channel (requires bot token)
./ogmo --channel telegram
# Expect: connects to Telegram, processes messages

# 3. Gateway service
./ogmo --gateway --port 8080
# Expect: HTTP server starts, health check responds
curl http://localhost:8080/health
# Expect: {"status":"ok"}
```

**Phase 8 milestone:** ogmo runs as a personal agent connected to IM channels.

---

## Summary

| Phase | Items | Description | Milestone | Manual Test |
|---|---|---|---|---|
| 1 | 13 | Foundation | Single streaming LLM call works | CLI flags, streaming response, context cancel |
| 2 | 11 | Safety & Prompts | Permissions, hooks, system prompt | permission prompts, plan mode, CLAUDE.md, hooks |
| 3 | 10 | Essential Tools | Basic coding agent usable | read/write/edit/bash/glob/grep workflows |
| 4 | 31 | Extended Tools | All 43 tools ported | plan mode, todos, cron, MCP, background tasks |
| 5 | 56 | UI & Commands | Full REPL with 54 slash commands | interactive REPL, all slash commands |
| 6 | 17 | Subsystems | Skills, memory, MCP, plugins, coordinator | skill loading, memory persistence, MCP connect, plugin install |
| 7 | 10 | API Providers | All provider types work | OpenAI provider, provider switching, auth |
| 8 | 16 | Channels (ogmo) | IM channel integrations | Telegram/Slack/Discord connect, gateway |
| **Total** | **164** | | | |

## Testing Conventions

### Unit Test Rules

Every ported item **must** have a `_test.go` file in the same package. Tests cover:

1. **Happy path** — valid input produces expected output
2. **Error cases** — missing input, invalid input, boundary conditions
3. **Concurrency** — goroutine safety where applicable (registry, manager)
4. **Context cancellation** — all I/O operations respect `ctx.Done()`

### Test Patterns

```go
// Use table-driven tests for multi-case scenarios
func TestChecker(t *testing.T) {
    tests := []struct {
        name     string
        mode     permissions.Mode
        tool     string
        want     permissions.Decision
    }{
        {"read_in_default", permissions.ModeDefault, "read_file", permissions.Allow},
        {"write_in_default", permissions.ModeDefault, "write_file", permissions.Ask},
        {"write_in_plan", permissions.ModePlan, "write_file", permissions.Deny},
        {"bash_in_auto", permissions.ModeAuto, "bash", permissions.Allow},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            c := permissions.NewDefaultChecker(tt.mode)
            got, _ := c.Check(context.Background(), permissions.Check{ToolName: tt.tool})
            assert.Equal(t, tt.want, got)
        })
    }
}

// Mock interfaces for isolated testing
type mockAPIClient struct {
    events []api.StreamEvent
    err    error
}
func (m *mockAPIClient) Stream(ctx context.Context, opts api.StreamOptions) (<-chan api.StreamEvent, error) {
    ch := make(chan api.StreamEvent, len(m.events))
    for _, e := range m.events { ch <- e }
    close(ch)
    return ch, m.err
}
```

### Running Tests

```bash
# All tests
go test ./...

# Single package
go test ./internal/engine/ -v

# Single test
go test ./internal/permissions/ -run TestChecker -v

# With coverage
go test -cover ./...

# Race detector
go test -race ./...
```

<!-- ## Priority Order for First Usable Build

The fastest path to a usable `og` binary:

```
1.1 → 1.2 → 1.3 → 1.4 → 1.5 → 1.6 → 1.7 → 1.8 → 1.10 → 1.11 → 1.13
                                                                  ↓
2.1 → 2.2 → 2.7 → 2.8 → 2.9 → 2.10
                                  ↓
3.1 → 3.2 → 3.3 → 3.4 → 3.5 → 3.6 → 3.9
                                            ↓
5.1 → 5.2 → 5.3 → 5.6 → 5.7 → 5.8
```

That's ~30 items for a working agent that can read, write, edit, and run commands. -->
