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

## Source Evidence

- `OpenHarness/src/openharness/settings.py` — environment variable handling
