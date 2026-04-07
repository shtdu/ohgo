# REQ-CF-003: Provider Profiles

**Pattern:** Ubiquitous
**Capability:** Configuration

## Requirement

The system shall support multiple provider profiles, each defining API connection parameters (base URL, API key, format) for different AI backends.

## Acceptance Criteria

- [ ] Multiple profiles can be defined (Anthropic, OpenAI, Copilot, etc.)
- [ ] Each profile specifies API format, base URL, and authentication
- [ ] Profiles are managed via `provider` subcommands (list, use, add, remove)

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `provider` subcommand
- `OpenHarness/src/openharness/settings.py` — profile storage
