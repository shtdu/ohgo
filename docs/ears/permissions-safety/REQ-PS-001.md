# REQ-PS-001: Permission Mode Enforcement

**Pattern:** Ubiquitous
**Capability:** Permissions and Safety

## Requirement

The system shall enforce tool execution permissions according to the active permission mode (default, plan, or full_auto).

## Acceptance Criteria

- [ ] Three modes are available: default, plan, full_auto
- [ ] The active mode is selectable via CLI flag or settings
- [ ] Every tool execution passes through the permission checker

## Source Evidence

- `OpenHarness/src/openharness/permissions/` — PermissionChecker, PermissionMode enum
- `OpenHarness/src/openharness/cli.py` — `--permission-mode` flag
