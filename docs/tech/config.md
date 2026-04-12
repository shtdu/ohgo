# Configuration Design

How ohgo discovers, loads, merges, and resolves configuration.

## Design Goals

1. **Zero-config defaults** — `og` works out of the box with a single `ANTHROPIC_API_KEY`
2. **Layered override** — every setting can be overridden at a more specific level
3. **Profile-driven** — users select a provider profile, not individual connection fields
4. **Python-compatible** — shared config directory and settings format with OpenHarness

## Config Layers

Settings are loaded from four layers, merged in order. Later layers override earlier ones:

```
defaults  →  user settings  →  project settings  →  environment variables
   │              │                  │                      │
   │              │                  │                      │
   ▼              ▼                  ▼                      ▼
 hardcoded    ~/.openharness/    ./.openharness/         ANTHROPIC_*,
 values       settings.json     settings.json            OPENHARNESS_*
```

CLI flags are applied after all layers, overriding everything.

### Merge Semantics

The merge is **field-level, not document-level**. Each non-zero field in the override replaces the corresponding field in the base. Maps (profiles) are merged by key — user profiles supplement the built-in catalog without replacing it.

Key rules:
- Zero-value fields (`""`, `0`, `false`, `nil`) are treated as "not set" and don't override
- Boolean fields use presence-as-true (setting `vim_mode` to any value in JSON enables it)
- Profile maps are merged: built-in profiles provide defaults, user profiles override by name
- Environment variables always win over file-based config

## Directory Resolution

The config directory resolves in order:

1. `OPENHARNESS_CONFIG_DIR` environment variable
2. `~/.openharness/` (default)

All paths derive from the config directory:

```
~/.openharness/
  settings.json        # user settings
  credentials.json     # auth credentials (file-based keyring)
  data/
    sessions/          # session history
    tasks/             # background task output logs
    feedback/          # user feedback storage
    cron_jobs.json     # scheduled job registry
  logs/                # application logs

./.openharness/
  settings.json        # project settings
  memory/              # project-local memory store
  plugins/             # project-local plugins
```

## Provider Profiles

Profiles abstract away the connection details for different LLM providers. Instead of configuring `api_format`, `auth_source`, `base_url`, and `default_model` individually, users select a profile by name.

### Built-in Profiles

| Profile | Provider | Auth Source | API Format | Default Model |
|---|---|---|---|---|
| `claude-api` | anthropic | `anthropic_api_key` | anthropic | claude-sonnet-4-6 |
| `claude-subscription` | anthropic (via CLI) | `claude_subscription` | anthropic | claude-sonnet-4-6 |
| `openai-compatible` | openai | `openai_api_key` | openai | gpt-5.4 |
| `codex` | openai (via CLI) | `codex_subscription` | openai | gpt-5.4 |
| `copilot` | copilot | `copilot_oauth` | copilot | gpt-5.4 |

### Profile Resolution

```
--profile flag
    │
    ▼
active_profile from settings
    │
    ▼
fallback to "claude-api"
    │
    ▼
merge with user profiles (user overrides supplement built-ins)
    │
    ▼
resolve auth → resolve API client → resolve model
```

Users can define custom profiles in `settings.json` to add third-party providers or override built-in profile defaults.

### Auth Resolution

Each profile declares an `auth_source` that determines how credentials are obtained:

| Auth Source | Mechanism | Env Key |
|---|---|---|
| `anthropic_api_key` | API key from env or credentials store | `ANTHROPIC_API_KEY` |
| `openai_api_key` | API key from env or credentials store | `OPENAI_API_KEY` |
| `claude_subscription` | Read from Claude CLI config | — |
| `codex_subscription` | Read from Codex CLI config | `CODEX_API_KEY` |
| `copilot_oauth` | GitHub OAuth device flow → Copilot token | `GITHUB_TOKEN` |

The auth manager tries sources in order: credentials store → environment variable → interactive flow (if applicable).

## Environment Variables

| Variable | Overrides |
|---|---|
| `ANTHROPIC_API_KEY` | API key |
| `ANTHROPIC_MODEL` | Default model |
| `ANTHROPIC_BASE_URL` | API base URL |
| `OPENHARNESS_MODEL` | Default model (fallback if ANTHROPIC_MODEL unset) |
| `OPENHARNESS_BASE_URL` | API base URL (fallback) |
| `OPENHARNESS_API_FORMAT` | API format (anthropic/openai/copilot) |
| `OPENHARNESS_PROVIDER` | Provider name |
| `OPENHARNESS_MAX_TOKENS` | Max response tokens |
| `OPENHARNESS_MAX_TURNS` | Max agent loop turns |
| `OPENHARNESS_CONFIG_DIR` | Config directory location |
| `OPENHARNESS_DATA_DIR` | Data directory location |
| `OPENHARNESS_LOGS_DIR` | Logs directory location |

`ANTHROPIC_*` variables take precedence over `OPENHARNESS_*` equivalents, supporting users who already have Anthropic SDK env vars set.

## Settings Schema

The settings file is JSON. Only non-default values need to be present — omitted fields inherit defaults.

### Connection

| Field | Default | Description |
|---|---|---|
| `api_key` | — | API key (prefer env var or credentials store) |
| `model` | `claude-sonnet-4-6` | Default model for queries |
| `max_tokens` | `16384` | Max response tokens per turn |
| `base_url` | provider default | Override API endpoint |
| `api_format` | `anthropic` | SSE protocol (anthropic/openai/copilot) |
| `provider` | — | Provider identifier |
| `active_profile` | `claude-api` | Selected provider profile |
| `max_turns` | `200` | Agent loop turn limit |

### Behavior

| Field | Default | Description |
|---|---|---|
| `system_prompt` | — | Appended to assembled system prompt |
| `permission.mode` | `default` | Permission mode (default/plan/auto) |
| `permission.allowed_tools` | `[]` | Tools that bypass the permission prompt |
| `permission.denied_tools` | `[]` | Tools that are always blocked |
| `permission.path_rules` | `[]` | Glob-based path allow/deny rules |
| `permission.denied_commands` | `[]` | Shell command deny patterns |
| `memory.enabled` | `true` | Enable cross-session memory |
| `memory.max_files` | `5` | Max memory entries per project |
| `memory.max_entrypoint_lines` | `200` | Max lines in MEMORY.md index |
| `mcp.servers` | `[]` | MCP server connection configs |

### UI

| Field | Default | Description |
|---|---|---|
| `theme` | `default` | Color theme |
| `output_style` | `default` | Output formatting style |
| `vim_mode` | `false` | Vim keybindings in REPL |
| `verbose` | `false` | Verbose logging |

## MCP Server Config

Each MCP server entry in `mcp.servers`:

| Field | Required | Description |
|---|---|---|
| `name` | yes | Unique server identifier |
| `transport` | yes | `stdio`, `sse`, or `streamable_http` |
| `command` | stdio only | Executable to launch |
| `args` | stdio only | Arguments to the command |
| `url` | sse/http only | Server URL |
| `headers` | no | HTTP headers for sse/http transports |
| `env` | no | Extra environment variables |
| `enabled` | no | Default `true`, set `false` to disable |

## Persistence

Settings are saved as indented JSON with a trailing newline. File permissions are `0600` to protect API keys. The `Save()` function writes to the user config path only — project config is read-only via the loading mechanism.
