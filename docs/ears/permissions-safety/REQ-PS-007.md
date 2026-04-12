# REQ-PS-007: Destructive Operation Warning

**Pattern:** Event-Driven
**Capability:** Permissions and Safety

## Requirement

When the agent attempts a destructive operation (e.g., force push, file deletion, database drop), the system shall block execution pending explicit user confirmation before proceeding.

## Acceptance Criteria

- [ ] Destructive patterns are detected in tool inputs using pattern matching
- [ ] Presents a warning containing the matched destructive pattern name and the specific file or command path
- [ ] Execution remains blocked until the user explicitly confirms or denies the operation
- [ ] When the user confirms, execution proceeds with the destructive operation; when denied, execution is cancelled

## Source Evidence

- `OpenHarness/src/openharness/permissions/` — destructive operation detection
