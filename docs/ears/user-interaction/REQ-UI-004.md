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

## Source Evidence

- `OpenHarness/src/openharness/commands/` — 54+ slash command implementations
