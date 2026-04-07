# REQ-PS-007: Destructive Operation Warning

**Pattern:** Event-Driven
**Capability:** Permissions and Safety

## Requirement

When the agent attempts a destructive operation (e.g., force push, file deletion, database drop), the system shall warn the user with a description of the operation's impact before execution.

## Acceptance Criteria

- [ ] Identifies known destructive patterns in tool inputs
- [ ] Presents a clear warning with the nature of the destructive action
- [ ] Requires explicit user confirmation to proceed

## Source Evidence

- `OpenHarness/src/openharness/permissions/` — destructive operation detection
