# Extensibility

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

---

# REQ-EX-002: Plugin Contribution Registration

**Pattern:** Event-Driven
**Capability:** Extensibility

## Requirement

When a plugin becomes active (first loaded per REQ-EX-001 or re-enabled per REQ-EX-008), the system shall register each of its contributions for use in the corresponding capability.

## Acceptance Criteria

- [ ] Plugin commands become available for invocation
- [ ] Plugin skills can be loaded on demand
- [ ] Plugin hooks execute before and after the associated tool invocation
- [ ] Plugin external tool servers are connected and their tools become available
- [ ] Contributions are registered when a plugin is first loaded (per REQ-EX-001) and when re-enabled (per REQ-EX-008)

## Source Evidence

- `OpenHarness/src/openharness/plugins/` — plugin contribution handling

---

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

---

# REQ-EX-004: On-Demand Skill Loading

**Pattern:** Event-Driven
**Capability:** Extensibility

## Requirement

When the agent invokes a skill, the system shall load the skill's markdown content and inject it into the agent's context for the current turn.

## Acceptance Criteria

- [ ] Skills are loaded from bundled, user, and plugin sources
- [ ] Each skill provides a name and description in its metadata
- [ ] The skill content becomes part of the agent's instructions for execution
- [ ] Skills are loaded on demand, not all at startup
- [ ] When a skill's markdown content is missing or unreadable, the system logs a warning and skips the skill
- [ ] When skill metadata (YAML frontmatter) is invalid or missing required fields, the system logs a warning with the skill name and skips the skill

## Source Evidence

- `OpenHarness/src/openharness/skills/` — skill registry and loading
- `OpenHarness/src/openharness/tools/skill_tool.py`

---

# REQ-EX-005: Hook Execution on Lifecycle Events

**Pattern:** Event-Driven
**Capability:** Extensibility

## Requirement

When a configured hook event fires (PreToolUse, PostToolUse), the system shall execute all registered hooks for that event in order, stopping early if any hook requests cancellation.

## Acceptance Criteria

- [ ] Hooks fire before and after tool execution
- [ ] Hooks execute in registration order
- [ ] A hook can alter the tool input or prevent the tool from executing
- [ ] A failing hook does not terminate the session; the tool execution proceeds unless the hook explicitly requests cancellation

## Source Evidence

- `OpenHarness/src/openharness/hooks/` — hook executor

---

# REQ-EX-006: Hook Type Support

**Pattern:** Optional Feature
**Capability:** Extensibility

## Requirement

Where a hook of a supported type is configured, the system shall execute it according to the deserialization and validation rules defined for its declared parameter type.

## Acceptance Criteria

- [ ] Command-type hooks execute a configured action and return its output
- [ ] Prompt-type hooks produce an AI-generated response
- [ ] URL-type hooks retrieve content from a web address
- [ ] Webhook-type hooks send a notification to an external service
- [ ] When a hook execution fails (command not found, network timeout), the system logs the error and continues the session without the hook result

## Source Evidence

- `OpenHarness/src/openharness/hooks/` — hook type implementations

---

# REQ-EX-007: MCP Server Management

**Pattern:** Event-Driven
**Capability:** Extensibility

## Requirement

When the user adds or removes an MCP server configuration, the system shall persist the configuration change and notify the runtime bridge (per REQ-TL-010).

## Acceptance Criteria

- [ ] MCP servers are added via CLI (`mcp add`) or settings
- [ ] MCP servers are removed via CLI (`mcp remove`)
- [ ] The configuration of external tool servers is persisted and reflected on next session start or MCP reconnection
- [ ] Server connection errors are reported without affecting other tools
- [ ] When adding an MCP server fails due to invalid configuration or connectivity issues, the system reports the error and does not add the server to the configuration

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `mcp` subcommand
- `OpenHarness/src/openharness/mcp/` — client manager

---

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

---
