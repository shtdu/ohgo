# REQ-AU-004: Authentication Status Reporting

**Pattern:** Event-Driven
**Capability:** Authentication

## Requirement

When the user checks authentication status (`auth status`), the system shall report the active provider and credential validity.

## Acceptance Criteria

- [ ] Shows the currently active provider profile
- [ ] Indicates whether credentials are valid or expired
- [ ] Lists available provider profiles

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `auth status` subcommand
