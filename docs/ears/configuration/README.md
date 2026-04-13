# Configuration

# REQ-CF-001: Settings File Configuration

**Pattern:** Ubiquitous
**Capability:** Configuration

## Requirement

The system shall read configuration from a JSON settings file at a standard location (`~/.openharness/settings.json`).

## Acceptance Criteria

- [ ] Reads settings from the default user config directory
- [ ] Supports an alternate settings file location when specified at startup
- [ ] The settings file schema covers all configuration parameters defined in the Configuration domain requirements (REQ-CF-002 through REQ-CF-007)
- [ ] When the settings file contains invalid JSON, the system reports a parse error with the file path and line number

## Source Evidence

- `OpenHarness/src/openharness/config/settings.py`
- `OpenHarness/src/openharness/cli.py` — `--settings` flag

---

# REQ-CF-002: CLI Flag Overrides

**Pattern:** Optional Feature
**Capability:** Configuration

## Requirement

Where CLI flags are provided at invocation (e.g., `--model`, `--permission-mode`, `--effort`, `--max-turns`), the system shall override the corresponding settings values for the duration of that session.

## Acceptance Criteria

- [ ] CLI flags override both global and project-level settings for the current session
- [ ] Flag values are not persisted to the settings file
- [ ] Unset flags fall through to the next configuration layer (project settings, then global settings)
- [ ] When a CLI flag value is invalid (e.g., non-existent model, unknown permission mode), the system reports the error before session start and exits with a non-zero status

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `--model`, `--permission-mode`, `--effort`, `--max-turns` flags

---

# REQ-CF-003: Provider Profiles

**Pattern:** Optional Feature
**Capability:** Configuration

## Requirement

Where provider profiles are defined, the system shall connect to AI backends using each profile's specified parameters (base URL, API key, format).

## Acceptance Criteria

- [ ] Multiple profiles can be defined (Anthropic, OpenAI, Copilot, etc.)
- [ ] Each profile specifies API format, base URL, and authentication
- [ ] The active profile determines which backend receives API requests; switching profiles changes the target backend
- [ ] When a profile has missing required fields or a duplicate name, the system rejects the profile and reports the specific validation error

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `provider` subcommand
- `OpenHarness/src/openharness/config/settings.py` — profile storage

---

# REQ-CF-004: Runtime Profile Switching

**Pattern:** Event-Driven
**Capability:** Configuration

## Requirement

When the user switches provider profiles, the system shall validate the new provider connection and update the active API configuration without restarting the session.

## Acceptance Criteria

- [ ] After a profile switch completes, the next API call uses the new provider configuration
- [ ] The active profile selection is persisted to the settings file so that future sessions use the same profile
- [ ] The system confirms the profile switch to the user, showing the new provider name

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `provider use` subcommand
- `/provider` slash command

---

# REQ-CF-005: Environment Variable Overrides

**Pattern:** Optional Feature
**Capability:** Configuration

## Requirement

Where environment variables are set (e.g., `ANTHROPIC_API_KEY`, `OPENHARNESS_MODEL`), the system shall use them as overrides for corresponding settings values.

## Acceptance Criteria

- [ ] `ANTHROPIC_API_KEY` / `OPENAI_API_KEY` override API key settings
- [ ] `OPENHARNESS_MODEL` overrides the default model
- [ ] `OPENHARNESS_SETTINGS` overrides the settings file path
- [ ] Environment variables take precedence over settings file but not CLI flags
- [ ] When an environment variable override contains an invalid value, the system reports which variable and the expected format

## Source Evidence

- `OpenHarness/src/openharness/config/settings.py` — environment variable handling

---

# REQ-CF-006: Runtime Configuration Updates

**Pattern:** Event-Driven
**Capability:** Configuration

## Requirement

When the user updates settings via slash commands (e.g., `/config`, `/model`, `/theme`), the system shall apply changes immediately without restart.

## Acceptance Criteria

- [ ] Settings changes take effect for subsequent operations
- [ ] Changes are persisted to the settings file
- [ ] The user is informed of the change
- [ ] When an invalid value is provided via slash command (e.g., non-existent model, invalid key-value pair), the system reports the specific error and does not apply the change

## Source Evidence

- `OpenHarness/src/openharness/commands/` — `/config`, `/model`, `/theme` commands

---

# REQ-CF-007: Multi-Layer Configuration Discovery

**Pattern:** Ubiquitous
**Capability:** Configuration

## Requirement

The system shall discover and merge configuration from multiple layers per session: user global settings, project local settings, and CLI overrides.

## Acceptance Criteria

- [ ] The system reads configuration from the user's global settings file at startup
- [ ] The system reads project-level settings from the project's configuration directory when present
- [ ] CLI flags override both layers
- [ ] When the same setting is defined in multiple layers, the value from the highest-precedence layer (CLI flags > project settings > global settings) is used
- [ ] When a configuration file is missing, unreadable, or contains invalid syntax, the system uses defaults for that layer and logs a warning

## Source Evidence

- `OpenHarness/src/openharness/config/settings.py` — multi-layer config loading

---
