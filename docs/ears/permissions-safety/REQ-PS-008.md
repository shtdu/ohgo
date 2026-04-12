# REQ-PS-008: Permission Error Fail-Safe

**Pattern:** Unwanted Behaviour
**Capability:** Permissions and Safety

## Requirement

If the permission system encounters an error during evaluation, the system shall refuse tool execution and report the error to the user.

## Acceptance Criteria

- [ ] Tool execution is blocked when permission checking fails
- [ ] The user receives an error message containing the tool name, the permission rule that failed, and the failure reason
- [ ] The error is logged with the tool name, permission context, and timestamp

## Source Evidence

- `OpenHarness/src/openharness/permissions/` — error handling in permission checker
