# REQ-PS-001: Permission Mode Enforcement

**Pattern:** State-Driven
**Capability:** Permissions and Safety

## Requirement

While a permission mode is active, the system shall enforce tool execution permissions according to that mode's rules.

## Acceptance Criteria

- [ ] The system provides selectable permission modes (default, plan, and full_auto; details per REQ-PS-002, REQ-PS-003, REQ-PS-004)
- [ ] Every tool invocation is routed through the permission system before execution
- [ ] When the permission mode changes during a session, the new mode's rules apply to all subsequent tool invocations
- [ ] When an invalid permission mode is specified, the system defaults to the most restrictive mode and logs a warning

## Source Evidence

- `OpenHarness/src/openharness/permissions/` — PermissionChecker, PermissionMode enum
- `OpenHarness/src/openharness/cli.py` — `--permission-mode` flag
