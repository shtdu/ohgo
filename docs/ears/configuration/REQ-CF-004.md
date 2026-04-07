# REQ-CF-004: Runtime Profile Switching

**Pattern:** Event-Driven
**Capability:** Configuration

## Requirement

When the user switches provider profiles, the system shall update the active API configuration without restarting the session.

## Acceptance Criteria

- [ ] Profile switching takes effect immediately
- [ ] Subsequent API calls use the new provider configuration
- [ ] The active profile is persisted for future sessions

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `provider use` subcommand
- `/provider` slash command
