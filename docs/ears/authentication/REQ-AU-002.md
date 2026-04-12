# REQ-AU-002: OAuth Device Flow

**Pattern:** Event-Driven
**Capability:** Authentication

## Requirement

When the user invokes an authentication command (login, status, logout, switch), the system shall perform the corresponding OAuth device flow or credential management operation.

## Acceptance Criteria

- [ ] Login displays a verification URL, user code, polls for token exchange, stores the resulting token with owner-only file permissions
- [ ] Status queries display the current authentication state
- [ ] Logout revokes the stored token
- [ ] Switch selects a different stored credential profile as the active profile
- [ ] When the OAuth device flow expires or is denied, the system returns an error and does not store credentials

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `auth` subcommand (login, status, logout, switch)
- `OpenHarness/src/openharness/auth/` — OAuth flow implementation
