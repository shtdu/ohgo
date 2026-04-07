# REQ-PS-003: Plan Mode Restrictions

**Pattern:** State-Driven
**Capability:** Permissions and Safety

## Requirement

While the system is in plan mode, the system shall restrict tool execution to read-only operations and planning tools.

## Acceptance Criteria

- [ ] File write, edit, and bash tools are disabled
- [ ] File read, search, and planning tools remain available
- [ ] Any attempt to invoke a write-capable tool while in plan mode is rejected with an informative message

## Source Evidence

- `OpenHarness/src/openharness/permissions/` — plan mode restrictions
- `OpenHarness/src/openharness/tools/enter_plan_mode_tool.py`
- `OpenHarness/src/openharness/tools/exit_plan_mode_tool.py`
