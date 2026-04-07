# REQ-CF-002: CLI Override Precedence

**Pattern:** Complex
**Capability:** Configuration

## Requirement

If a CLI flag overrides a settings value, then the system shall use the CLI value for the duration of the session and revert to the settings value on next launch.

## Acceptance Criteria

- [ ] CLI flags take precedence over settings file values
- [ ] Overrides apply only to the current session
- [ ] The settings file is not modified by CLI overrides

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — flag parsing and configuration assembly
