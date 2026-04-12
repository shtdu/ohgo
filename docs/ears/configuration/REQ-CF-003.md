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
