# REQ-CF-003: Provider Profiles

**Pattern:** Optional Feature
**Capability:** Configuration

## Requirement

Where provider profiles are defined, the system shall connect to AI backends using each profile's specified parameters (base URL, API key, format).

## Acceptance Criteria

- [ ] Multiple profiles can be defined (Anthropic, OpenAI, Copilot, etc.)
- [ ] Each profile specifies API format, base URL, and authentication
- [ ] Profiles are managed via `provider` subcommands (list, use, add, remove)

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `provider` subcommand
- `OpenHarness/src/openharness/settings.py` — profile storage
