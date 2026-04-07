# REQ-CF-001: Settings File Configuration

**Pattern:** Ubiquitous
**Capability:** Configuration

## Requirement

The system shall read configuration from a JSON settings file at a standard location (`~/.openharness/settings.json`).

## Acceptance Criteria

- [ ] Reads settings from the default user config directory
- [ ] Supports an alternate settings file via `--settings` flag
- [ ] Settings file contains all configurable parameters

## Source Evidence

- `OpenHarness/src/openharness/settings.py`
- `OpenHarness/src/openharness/cli.py` — `--settings` flag
