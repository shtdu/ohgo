# REQ-CF-004: Runtime Profile Switching

**Pattern:** Event-Driven
**Capability:** Configuration

## Requirement

When the user switches provider profiles, the system shall update the active API configuration without restarting the session. Profile switching is a specialized case of the runtime configuration update mechanism defined in REQ-CF-006, retained as a separate requirement for its distinct acceptance criteria around API connection continuity.

## Acceptance Criteria

- [ ] After a profile switch completes, the next API call uses the new provider configuration
- [ ] The active profile is persisted for future sessions
- [ ] The system confirms the profile switch to the user, showing the new provider name

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `provider use` subcommand
- `/provider` slash command
