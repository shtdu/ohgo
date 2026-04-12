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
