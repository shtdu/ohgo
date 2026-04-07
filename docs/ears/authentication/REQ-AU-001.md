# REQ-AU-001: API Key Authentication

**Pattern:** Ubiquitous
**Capability:** Authentication

## Requirement

The system shall authenticate with AI providers using API keys sourced from configuration, environment variables, or interactive input.

## Acceptance Criteria

- [ ] API keys are read from settings file
- [ ] API keys can be overridden by environment variables
- [ ] API keys can be passed via `--api-key` CLI flag
- [ ] Missing API keys produce a clear error message

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `--api-key` flag
- `OpenHarness/src/openharness/settings.py` — `api_key` setting
