# REQ-EX-008: Plugin Enable and Disable

**Pattern:** Event-Driven
**Capability:** Extensibility

## Requirement

When the user enables or disables a plugin, the system shall update the active plugin set without restarting the session.

## Acceptance Criteria

- [ ] Disabled plugins are skipped during discovery
- [ ] Enabling a plugin loads its contributions immediately
- [ ] Plugin enable state is persisted in settings
- [ ] When a plugin is disabled mid-session, its tools and commands are removed from active use

## Source Evidence

- `OpenHarness/src/openharness/config/settings.py` — `enabled_plugins` dictionary
