# REQ-CF-006: Runtime Configuration Updates

**Pattern:** Event-Driven
**Capability:** Configuration

## Requirement

When the user updates settings via slash commands (e.g., `/config`, `/model`, `/theme`), the system shall apply changes immediately without restart.

## Acceptance Criteria

- [ ] Settings changes take effect for subsequent operations
- [ ] Changes are persisted to the settings file
- [ ] The user is informed of the change

## Source Evidence

- `OpenHarness/src/openharness/commands/` — `/config`, `/model`, `/theme` commands
