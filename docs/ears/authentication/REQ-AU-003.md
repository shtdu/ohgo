# REQ-AU-003: Multi-Provider Backend Support

**Pattern:** State-Driven
**Capability:** Authentication

## Requirement

While a provider profile is active, the system shall authenticate with that provider using the credentials and protocol specified in the profile (per Configuration domain).

## Acceptance Criteria

- [ ] When a standard API key provider is active, the system uses the API key and base URL from the profile for authentication
- [ ] When an OAuth-based provider is active, the system uses the stored OAuth token (per REQ-AU-002)
- [ ] When a subscription bridge provider is active, the system delegates credential verification to the associated CLI tool and reports success or failure to the user
- [ ] When no provider profile is active, the system produces an error indicating that a provider must be configured

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `--api-format` flag
- `OpenHarness/src/openharness/bridge/` — subscription bridge implementations
