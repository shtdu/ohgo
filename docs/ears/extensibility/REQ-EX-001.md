# REQ-EX-001: Plugin Discovery and Loading

**Pattern:** Optional Feature
**Capability:** Extensibility

## Requirement

Where plugin directories are configured, the system shall discover and load plugins from those directories, reading each plugin's manifest to determine its contributions.

## Acceptance Criteria

- [ ] Discovers plugins from all configured directories (user and project scope)
- [ ] Each plugin provides a manifest declaring its contributions
- [ ] Skips plugins with invalid manifests and reports the error
- [ ] When the plugin directory is missing or inaccessible, the system logs a warning and continues with built-in capabilities

## Source Evidence

- `OpenHarness/src/openharness/plugins/` — plugin discovery and loading
