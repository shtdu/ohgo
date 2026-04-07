# REQ-PS-002: Default Mode User Confirmation

**Pattern:** Event-Driven
**Capability:** Permissions and Safety

## Requirement

When the permission mode is default and the agent requests a sensitive tool execution, the system shall prompt the user for approval before proceeding.

## Acceptance Criteria

- [ ] Sensitive tools (file write, bash, etc.) require user approval
- [ ] Read-only tools may bypass confirmation based on configuration
- [ ] The user can approve, deny, or approve-all for the session

## Source Evidence

- `OpenHarness/src/openharness/permissions/` — default mode behavior
