# REQ-EX-005: Hook Execution on Lifecycle Events

**Pattern:** Event-Driven
**Capability:** Extensibility

## Requirement

When a configured hook event fires (PreToolUse, PostToolUse), the system shall execute all registered hooks for that event in order.

## Acceptance Criteria

- [ ] Hooks fire before and after tool execution
- [ ] Hooks execute in registration order
- [ ] A hook can alter the tool input or prevent the tool from executing
- [ ] A failing hook does not terminate the session; the tool execution proceeds unless the hook explicitly requests cancellation

## Source Evidence

- `OpenHarness/src/openharness/hooks/` — hook executor
