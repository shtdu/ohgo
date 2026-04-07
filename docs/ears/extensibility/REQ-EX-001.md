# REQ-EX-001: Plugin Discovery and Loading

**Pattern:** Ubiquitous
**Capability:** Extensibility

## Requirement

The system shall discover and load plugins from configured plugin directories, reading each plugin's manifest to determine its contributions.

## Acceptance Criteria

- [ ] Discovers plugins from user directory (`~/.ohmo/plugins/`)
- [ ] Discovers plugins from project directory (`.openharness/plugins/`)
- [ ] Reads `plugin.json` manifest from each plugin directory
- [ ] Skips plugins with invalid manifests and logs the error

## Source Evidence

- `OpenHarness/src/openharness/plugins/` — plugin discovery and loading
