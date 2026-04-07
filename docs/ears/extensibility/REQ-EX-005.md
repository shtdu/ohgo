# REQ-EX-005: Hook Execution on Lifecycle Events

**Pattern:** Event-Driven
**Capability:** Extensibility

## Requirement

When a configured hook event fires (PreToolUse, PostToolUse), the system shall execute all registered hooks for that event in order.

## Acceptance Criteria

- [ ] Hooks fire before and after tool execution
- [ ] Hooks execute in registration order
- [ ] Hook output can modify or block tool execution
- [ ] Hook failures are logged without crashing the agent

## Source Evidence

- `OpenHarness/src/openharness/hooks/` — hook executor
