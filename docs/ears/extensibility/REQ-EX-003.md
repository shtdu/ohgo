# REQ-EX-003: Plugin Lifecycle Management

**Pattern:** Event-Driven
**Capability:** Extensibility

## Requirement

When a plugin is installed, uninstalled, or listed via the CLI, the system shall update the plugin registry and reload affected subsystems.

## Acceptance Criteria

- [ ] `plugin install` registers a new plugin
- [ ] `plugin uninstall` removes a plugin
- [ ] `plugin list` shows installed plugins and status
- [ ] After install, the plugin's contributions become available
- [ ] When plugin installation fails (invalid archive, dependency failure), the system reports the specific error and does not register the plugin

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `plugin` subcommand
