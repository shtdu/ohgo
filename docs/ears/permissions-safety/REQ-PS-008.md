# REQ-PS-008: Permission Error Fail-Safe

**Pattern:** Unwanted Behaviour
**Capability:** Permissions and Safety

## Requirement

If the permission system encounters an error during evaluation, the system shall not execute the requested tool and shall report the error to the user.

## Acceptance Criteria

- [ ] Tool execution is blocked when permission checking fails
- [ ] The user receives a descriptive error message
- [ ] The error is logged for diagnostics

## Source Evidence

- `OpenHarness/src/openharness/permissions/` — error handling in permission checker
