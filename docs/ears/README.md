# OpenHarness Product Requirements

## Product Scope

**System:** OpenHarness — an AI agent harness providing tool-use, skills, memory, permissions, multi-agent coordination, and MCP integration via CLI, TUI, and messaging channel interfaces.

**Source:** Analysis of the OpenHarness Python codebase at `OpenHarness/`.

**Version baseline:** Current main branch.

## Capability Tree (MECE)

```
OpenHarness
├── User Interaction          # How users communicate with the system
│   ├── CLI interface         # REQ-UI-001, REQ-UI-002
│   ├── Terminal UI           # REQ-UI-003, REQ-UI-007, REQ-UI-008
│   ├── Slash commands        # REQ-UI-004
│   ├── Channel gateways      # REQ-UI-005
│   └── Interactive prompts   # REQ-UI-006
├── Tool Execution            # How tools are registered, found, executed
│   ├── Tool registry         # REQ-TL-001
│   ├── File operations       # REQ-TL-002
│   ├── Shell execution       # REQ-TL-003, REQ-TL-012
│   ├── Search tools          # REQ-TL-004, REQ-TL-005
│   ├── Web tools             # REQ-TL-006, REQ-TL-007
│   ├── Dev tools             # REQ-TL-008, REQ-TL-009
│   ├── MCP bridge            # REQ-TL-010
│   └── Tool discovery        # REQ-TL-011
├── Agent Coordination        # How multiple agents collaborate
│   ├── Subagent spawning     # REQ-AC-001, REQ-AC-004
│   ├── Team management       # REQ-AC-002
│   ├── Inter-agent messaging # REQ-AC-003
│   └── Task lifecycle        # REQ-AC-005
├── Memory and Context        # How information persists across sessions
│   ├── Persistent memory     # REQ-MC-001, REQ-MC-005, REQ-MC-006
│   ├── Context loading       # REQ-MC-002, REQ-MC-003
│   └── Memory search         # REQ-MC-004
├── Session Management        # How conversations are managed
│   ├── Persistence           # REQ-SM-001
│   ├── Continue/Resume       # REQ-SM-002, REQ-SM-003
│   ├── Export/Share          # REQ-SM-004, REQ-SM-005
│   ├── Tagging/Rewind        # REQ-SM-006, REQ-SM-007
│   └── Context compaction    # REQ-SM-008
├── Permissions and Safety    # How access is controlled
│   ├── Permission modes      # REQ-PS-001, REQ-PS-002, REQ-PS-003, REQ-PS-004
│   ├── Tool lists            # REQ-PS-005
│   ├── Path rules            # REQ-PS-006
│   ├── Destructive warnings  # REQ-PS-007
│   └── Fail-safe             # REQ-PS-008
├── Configuration             # How the system is customized
│   ├── Settings file         # REQ-CF-001, REQ-CF-007
│   ├── CLI overrides         # REQ-CF-002, REQ-CF-005
│   ├── Provider profiles     # REQ-CF-003, REQ-CF-004
│   └── Runtime updates       # REQ-CF-006
├── Extensibility             # How the system is extended
│   ├── Plugin system         # REQ-EX-001, REQ-EX-002, REQ-EX-003, REQ-EX-008
│   ├── Skills                # REQ-EX-004
│   ├── Hooks                 # REQ-EX-005, REQ-EX-006
│   └── MCP management        # REQ-EX-007
├── Task Automation           # How tasks are automated
│   ├── Background tasks      # REQ-AT-001, REQ-AT-002, REQ-AT-004, REQ-AT-005
│   └── Cron scheduling       # REQ-AT-003
└── Authentication            # How credentials are managed
    ├── API key auth          # REQ-AU-001
    ├── OAuth flow            # REQ-AU-002
    ├── Multi-provider        # REQ-AU-003
    └── Status reporting      # REQ-AU-004
```

## EARS Pattern Summary

| Pattern | Count | Percentage |
|---------|-------|------------|
| Ubiquitous | 17 | 25% |
| Event-Driven | 35 | 51% |
| State-Driven | 4 | 6% |
| Optional Feature | 8 | 12% |
| Complex | 5 | 7% |
| Unwanted Behaviour | 1 | 1% |
| **Total** | **70** | **100%** |

## Traceability

| Source | Requirements |
|--------|-------------|
| `OpenHarness/src/openharness/cli.py` | REQ-UI-001, REQ-UI-002, REQ-UI-007, REQ-SM-002, REQ-SM-003, REQ-PS-001, REQ-PS-005, REQ-CF-001, REQ-CF-002, REQ-CF-003, REQ-CF-004, REQ-AT-003, REQ-AU-001, REQ-AU-002, REQ-AU-004 |
| `OpenHarness/src/openharness/tools/` | REQ-TL-001..REQ-TL-012, REQ-AC-001..REQ-AC-005, REQ-AT-001..REQ-AT-005 |
| `OpenHarness/src/openharness/permissions/` | REQ-PS-001..REQ-PS-008 |
| `OpenHarness/src/openharness/memory/` | REQ-MC-001..REQ-MC-006 |
| `OpenHarness/src/openharness/commands/` | REQ-UI-004, REQ-UI-008, REQ-SM-004..REQ-SM-008, REQ-CF-006 |
| `OpenHarness/src/openharness/plugins/` | REQ-EX-001..REQ-EX-003, REQ-EX-008 |
| `OpenHarness/src/openharness/skills/` | REQ-EX-004 |
| `OpenHarness/src/openharness/hooks/` | REQ-EX-005, REQ-EX-006 |
| `OpenHarness/src/openharness/mcp/` | REQ-TL-010, REQ-EX-007 |
| `OpenHarness/src/openharness/settings.py` | REQ-CF-001, REQ-CF-005, REQ-CF-007, REQ-PS-005, REQ-PS-006, REQ-MC-006, REQ-AU-001 |
| `OpenHarness/src/openharness/auth/` | REQ-AU-002 |
| `OpenHarness/src/openharness/bridge/` | REQ-AU-003 |
| `OpenHarness/src/openharness/channels/` | REQ-UI-005 |
| `OpenHarness/src/openharness/swarm/` | REQ-AC-001, REQ-AC-003, REQ-AC-004 |
| `OpenHarness/src/openharness/tui/` | REQ-UI-003 |
| `OpenHarness/src/openharness/prompts/` | REQ-MC-003 |

## CE Check — Entry Point Coverage

### CLI Flags
- [x] `--version` / `-v` → REQ-UI-001
- [x] `--model` / `-m` → REQ-UI-002, REQ-CF-006
- [x] `--permission-mode` → REQ-PS-001, REQ-PS-002, REQ-PS-003, REQ-PS-004
- [x] `--effort` → REQ-UI-002
- [x] `--output-format` → REQ-UI-002
- [x] `--print` / `-p` → REQ-UI-002
- [x] `--max-turns` → REQ-UI-002
- [x] `--continue` / `-c` → REQ-SM-002
- [x] `--resume` / `-r` → REQ-SM-003
- [x] `--settings` → REQ-CF-001
- [x] `--system-prompt` / `-s` → REQ-UI-002
- [x] `--api-key` / `-k` → REQ-AU-001
- [x] `--api-format` → REQ-AU-003
- [x] `--base-url` → REQ-CF-003
- [x] `--theme` → REQ-UI-007
- [x] `--bare` → REQ-CF-007
- [x] `--allowed-tools` → REQ-PS-005
- [x] `--disallowed-tools` → REQ-PS-005
- [x] `--dangerously-skip-permissions` → REQ-PS-001

### CLI Subcommands
- [x] `mcp add/remove/list` → REQ-EX-007
- [x] `plugin install/uninstall/list` → REQ-EX-003
- [x] `auth login/status/logout/switch` → REQ-AU-002, REQ-AU-004
- [x] `provider list/use/add/remove` → REQ-CF-003, REQ-CF-004
- [x] `cron start/stop/status/list/toggle` → REQ-AT-003

### Slash Commands
- [x] `/help`, `/exit`, `/clear` → REQ-UI-004
- [x] `/commit`, `/debug`, `/plan`, `/review`, `/test` → REQ-UI-004 (bundled skills via REQ-EX-004)
- [x] `/resume`, `/continue` → REQ-SM-002, REQ-SM-003
- [x] `/export`, `/share` → REQ-SM-004, REQ-SM-005
- [x] `/tag`, `/rewind` → REQ-SM-006, REQ-SM-007
- [x] `/compact` → REQ-SM-008
- [x] `/config` → REQ-CF-006
- [x] `/model`, `/theme`, `/vim`, `/fast`, `/effort` → REQ-CF-006
- [x] `/permissions`, `/plan` → REQ-PS-001, REQ-PS-003
- [x] `/provider` → REQ-CF-004
- [x] `/tasks`, `/agents` → REQ-AT-001, REQ-AT-002
- [x] `/mcp`, `/plugin`, `/reload-plugins` → REQ-EX-007, REQ-EX-003
- [x] `/memory` → REQ-MC-005
- [x] `/skills` → REQ-EX-004
- [x] `/hooks` → REQ-EX-005
- [x] `/session` → REQ-SM-001

### User Actions
- [x] Ask natural language prompt → REQ-UI-001
- [x] Approve/deny tool execution → REQ-PS-002
- [x] Switch permission mode → REQ-PS-001
- [x] Start/continue/resume session → REQ-SM-001, REQ-SM-002, REQ-SM-003
- [x] Add/remove memory → REQ-MC-005
- [x] Create background task → REQ-AT-001
- [x] Spawn subagent → REQ-AC-001
- [x] Create/manage team → REQ-AC-002
- [x] Install/uninstall plugin → REQ-EX-003
- [x] Add/remove MCP server → REQ-EX-007
- [x] Login/logout auth → REQ-AU-002
- [x] Switch provider profile → REQ-CF-004

### Configuration Keys
- [x] `api_key` → REQ-AU-001
- [x] `model` → REQ-CF-006
- [x] `permission.mode` → REQ-PS-001
- [x] `permission.allowed_tools` → REQ-PS-005
- [x] `permission.denied_tools` → REQ-PS-005
- [x] `permission.path_rules` → REQ-PS-006
- [x] `memory.enabled` → REQ-MC-001
- [x] `memory.max_files` → REQ-MC-006
- [x] `mcp_servers` → REQ-EX-007
- [x] `enabled_plugins` → REQ-EX-008
- [x] `theme` → REQ-UI-007
- [x] `vim_mode` → REQ-UI-008

### External Integrations
- [x] Anthropic API → REQ-AU-003
- [x] OpenAI-compatible API → REQ-AU-003
- [x] GitHub Copilot (OAuth) → REQ-AU-002, REQ-AU-003
- [x] Subscription bridges (Claude CLI, Codex CLI) → REQ-AU-003
- [x] MCP servers → REQ-TL-010, REQ-EX-007
- [x] LSP servers → REQ-TL-008
- [x] Channel gateways (Telegram, Slack, Discord, Feishu, etc.) → REQ-UI-005
- [x] Web fetch/search → REQ-TL-006, REQ-TL-007

### Tools (43+)
- [x] bash → REQ-TL-003, REQ-TL-012
- [x] file_read → REQ-TL-002
- [x] file_write → REQ-TL-002
- [x] file_edit → REQ-TL-002
- [x] notebook_edit → REQ-TL-009
- [x] glob → REQ-TL-004
- [x] grep → REQ-TL-005
- [x] web_fetch → REQ-TL-006
- [x] web_search → REQ-TL-007
- [x] lsp → REQ-TL-008
- [x] ask_user_question → REQ-UI-006
- [x] skill → REQ-EX-004
- [x] tool_search → REQ-TL-011
- [x] config → REQ-CF-006
- [x] enter_plan_mode / exit_plan_mode → REQ-PS-003
- [x] enter_worktree / exit_worktree → REQ-UI-002 (dev tools)
- [x] task_create/get/list/output/stop/update → REQ-AT-001..REQ-AT-005
- [x] team_create/delete, send_message → REQ-AC-002, REQ-AC-003
- [x] agent → REQ-AC-001
- [x] cron_create/delete/list/toggle, remote_trigger → REQ-AT-003
- [x] list_mcp_resources, read_mcp_resource, mcp_auth → REQ-TL-010
- [x] todo_write, brief → REQ-UI-002 (planning tools)
- [x] sleep → (utility, covered by REQ-TL-001 general registry)

**Gaps: None found.** All entry points are covered by at least one requirement.
