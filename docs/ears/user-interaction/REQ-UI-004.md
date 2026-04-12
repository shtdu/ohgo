# REQ-UI-004: Slash Commands

**Pattern:** Event-Driven
**Capability:** User Interaction

## Requirement

When a user enters a slash command (e.g., `/help`, `/commit`, `/plan`), the system shall execute the corresponding built-in or plugin-registered command.

## Acceptance Criteria

- [ ] The system recognizes commands starting with `/`
- [ ] Built-in commands include at minimum: help, exit, clear, commit, plan, status, config
- [ ] Plugins can register additional slash commands
- [ ] Unknown commands produce a descriptive error message
- [ ] When a built-in command fails during execution, the system reports the error and returns to the prompt loop
- [ ] When a plugin registers a command name matching a built-in command, the plugin command is namespaced and does not override the built-in

## Source Evidence

- `OpenHarness/src/openharness/commands/` — 54+ slash command implementations
