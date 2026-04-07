# REQ-PS-004: Full Auto Mode Execution

**Pattern:** State-Driven
**Capability:** Permissions and Safety

## Requirement

While the system is in full auto mode, the system shall execute tools without user confirmation within configured boundaries.

## Acceptance Criteria

- [ ] Tools execute without user confirmation
- [ ] Denied tools list still blocks execution
- [ ] Path rules still restrict file operations
- [ ] Destructive operation warnings (per REQ-PS-007) remain active even in full auto mode

## Source Evidence

- `OpenHarness/src/openharness/permissions/` — full_auto mode behavior
