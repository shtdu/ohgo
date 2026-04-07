# REQ-CF-001: Settings File Configuration

**Pattern:** Ubiquitous
**Capability:** Configuration

## Requirement

The system shall read configuration from a JSON settings file at a standard location (`~/.openharness/settings.json`).

## Acceptance Criteria

- [ ] Reads settings from the default user config directory
- [ ] Supports an alternate settings file location when specified at startup
- [ ] The settings file schema covers all configuration parameters defined in the Configuration domain requirements (REQ-CF-001 through REQ-CF-007)

## Source Evidence

- `OpenHarness/src/openharness/settings.py`
- `OpenHarness/src/openharness/cli.py` — `--settings` flag
