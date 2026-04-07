# REQ-EX-001: Plugin Discovery and Loading

**Pattern:** Optional Feature
**Capability:** Extensibility

## Requirement

Where plugin directories are configured, the system shall discover and load plugins from those directories, reading each plugin's manifest to determine its contributions.

## Acceptance Criteria

- [ ] Discovers plugins from user directory (`~/.ohmo/plugins/`)
- [ ] Discovers plugins from project directory (`.openharness/plugins/`)
- [ ] Each plugin provides a manifest declaring its contributions
- [ ] Skips plugins with invalid manifests and reports the error

## Source Evidence

- `OpenHarness/src/openharness/plugins/` — plugin discovery and loading
