# REQ-AU-001: API Key Authentication

**Pattern:** Optional Feature
**Capability:** Authentication

## Requirement

Where an AI provider is configured, the system shall authenticate using API keys sourced from configuration, environment variables, or interactive input.

## Acceptance Criteria

- [ ] When a valid API key is provided through any source, the system successfully completes an authenticated request to the provider
- [ ] API keys are sourced from configuration (per Configuration domain), environment variables (per REQ-CF-005), or interactive input, in that precedence order
- [ ] When no API key is found through any source, the system produces an error indicating the missing credential and the expected configuration key
- [ ] When an API key is invalid or expired, the system returns an authentication error from the provider without exposing the key value

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `--api-key` flag
- `OpenHarness/src/openharness/config/settings.py` — `api_key` setting
