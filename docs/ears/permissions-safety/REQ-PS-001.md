# REQ-PS-001: Permission Mode Enforcement

**Pattern:** State-Driven
**Capability:** Permissions and Safety

## Requirement

While a permission mode is active, the system shall enforce tool execution permissions according to that mode's rules.

## Acceptance Criteria

- [ ] The system provides selectable permission modes (default, plan, and full_auto; details per REQ-PS-002, REQ-PS-003, REQ-PS-004)
- [ ] Every tool invocation is routed through the permission system before execution
- [ ] The active permission mode can be changed during a session via slash command or settings update

## Source Evidence

- `OpenHarness/src/openharness/permissions/` — PermissionChecker, PermissionMode enum
- `OpenHarness/src/openharness/cli.py` — `--permission-mode` flag
