# REQ-AU-002: OAuth Device Flow

**Pattern:** Event-Driven
**Capability:** Authentication

## Requirement

When the user initiates authentication with an OAuth provider (e.g., GitHub Copilot), the system shall perform the device authorization flow including user code display and token exchange.

## Acceptance Criteria

- [ ] Displays a verification URL and user code
- [ ] Polls for token exchange completion
- [ ] Stores the resulting token securely
- [ ] Supports login, status check, logout, and account switching

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `auth` subcommand (login, status, logout, switch)
- `OpenHarness/src/openharness/auth/` — OAuth flow implementation
