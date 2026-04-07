# REQ-CF-007: Multi-Layer Configuration Discovery

**Pattern:** Ubiquitous
**Capability:** Configuration

## Requirement

The system shall discover and merge configuration from multiple layers: user global settings, project local settings, and CLI overrides.

## Acceptance Criteria

- [ ] The system reads configuration from the user's global settings file at startup
- [ ] The system reads project-level settings from the project's configuration directory when present
- [ ] CLI flags override both layers
- [ ] When the same setting is defined in multiple layers, the value from the highest-precedence layer (CLI flags > project settings > global settings) is used
- [ ] CLI flag overrides apply only to the current session; the settings file is not modified
- [ ] On next launch, settings revert to values from the settings file

## Source Evidence

- `OpenHarness/src/openharness/settings.py` — multi-layer config loading
