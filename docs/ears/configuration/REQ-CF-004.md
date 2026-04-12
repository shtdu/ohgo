# REQ-CF-004: Runtime Profile Switching

**Pattern:** Event-Driven
**Capability:** Configuration

## Requirement

When the user switches provider profiles, the system shall validate the new provider connection and update the active API configuration without restarting the session.

## Acceptance Criteria

- [ ] After a profile switch completes, the next API call uses the new provider configuration
- [ ] The active profile selection is persisted to the settings file so that future sessions use the same profile
- [ ] The system confirms the profile switch to the user, showing the new provider name

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `provider use` subcommand
- `/provider` slash command
