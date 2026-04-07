# REQ-PS-002: Default Mode User Confirmation

**Pattern:** Event-Driven
**Capability:** Permissions and Safety

## Requirement

When the permission mode is default and the agent requests a sensitive tool execution, the system shall prompt the user for approval before proceeding.

## Acceptance Criteria

- [ ] Tools classified as write-capable (file write, file edit, command execution) require user approval
- [ ] Read-only tools bypass confirmation when that option is enabled in settings; otherwise all tools require confirmation
- [ ] The user can approve a single action, deny it, or approve all remaining actions for the session

## Source Evidence

- `OpenHarness/src/openharness/permissions/` — default mode behavior
