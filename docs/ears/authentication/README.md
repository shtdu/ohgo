# Authentication

How the system authenticates with AI providers — API keys, OAuth device flow, subscription bridges, and credential management.

## Requirements

| ID | Title | Pattern |
|----|-------|---------|
| [REQ-AU-001](#req-au-001-api-key-authentication) | API Key Authentication | Optional Feature |
| [REQ-AU-002](#req-au-002-oauth-device-flow) | OAuth Device Flow | Event-Driven |
| [REQ-AU-003](#req-au-003-multi-provider-backend-support) | Multi-Provider Backend Support | State-Driven |
| [REQ-AU-004](#req-au-004-authentication-status-reporting) | Authentication Status Reporting | Event-Driven |

## Dependencies

Cross-references to other domains:
- [Configuration](../configuration/README.md)

## Details

## REQ-AU-001: API Key Authentication

**Pattern:** Optional Feature

### Requirement

Where an AI provider is configured, the system shall authenticate using API keys sourced from configuration, environment variables, or interactive input.

### Acceptance Criteria

- [ ] When a valid API key is provided through any source, the system successfully completes an authenticated request to the provider
- [ ] API keys are sourced from configuration (per Configuration domain), environment variables (per REQ-CF-005), or interactive input, in that precedence order
- [ ] When no API key is found through any source, the system produces an error indicating the missing credential and the expected configuration key
- [ ] When an API key is invalid or expired, the system returns an authentication error from the provider without exposing the key value

### Source Evidence

- `OpenHarness/src/openharness/cli.py` — `--api-key` flag
- `OpenHarness/src/openharness/config/settings.py` — `api_key` setting


---

## REQ-AU-002: OAuth Device Flow

**Pattern:** Event-Driven

### Requirement

When the user invokes an authentication command (login, status, logout, switch), the system shall perform the corresponding OAuth device flow or credential management operation.

### Acceptance Criteria

- [ ] Login displays a verification URL, user code, polls for token exchange, stores the resulting token with owner-only file permissions
- [ ] Status queries display the current authentication state
- [ ] Logout revokes the stored token
- [ ] Switch selects a different stored credential profile as the active profile
- [ ] When the OAuth device flow expires or is denied, the system returns an error and does not store credentials

### Source Evidence

- `OpenHarness/src/openharness/cli.py` — `auth` subcommand (login, status, logout, switch)
- `OpenHarness/src/openharness/auth/` — OAuth flow implementation


---

## REQ-AU-003: Multi-Provider Backend Support

**Pattern:** State-Driven

### Requirement

While a provider profile is active, the system shall authenticate with that provider using the credentials and protocol specified in the profile (per Configuration domain).

### Acceptance Criteria

- [ ] When a standard API key provider is active, the system uses the API key and base URL from the profile for authentication
- [ ] When an OAuth-based provider is active, the system uses the stored OAuth token (per REQ-AU-002)
- [ ] When a subscription bridge provider is active, the system delegates credential verification to the associated CLI tool and reports success or failure to the user
- [ ] When no provider profile is active, the system produces an error indicating that a provider must be configured

### Source Evidence

- `OpenHarness/src/openharness/cli.py` — `--api-format` flag
- `OpenHarness/src/openharness/bridge/` — subscription bridge implementations


---

## REQ-AU-004: Authentication Status Reporting

**Pattern:** Event-Driven

### Requirement

When the user checks authentication status (`auth status`), the system shall report the active provider and credential validity.

### Acceptance Criteria

- [ ] Shows the currently active provider profile
- [ ] Indicates whether credentials are valid or expired
- [ ] Lists available provider profiles

### Source Evidence

- `OpenHarness/src/openharness/cli.py` — `auth status` subcommand
