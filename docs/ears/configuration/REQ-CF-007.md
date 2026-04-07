# REQ-CF-007: Multi-Layer Configuration Discovery

**Pattern:** Ubiquitous
**Capability:** Configuration

## Requirement

The system shall discover and merge configuration from multiple layers: user global settings, project local settings, and CLI overrides.

## Acceptance Criteria

- [ ] Global settings from `~/.openharness/settings.json`
- [ ] Project settings from `.openharness/` in the working directory
- [ ] CLI flags override both layers
- [ ] Layer precedence is consistent and documented

## Source Evidence

- `OpenHarness/src/openharness/settings.py` — multi-layer config loading
