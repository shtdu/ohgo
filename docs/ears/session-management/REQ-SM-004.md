# REQ-SM-004: Session Export

**Pattern:** Event-Driven
**Capability:** Session Management

## Requirement

When the user requests an export (`/export`), the system shall produce a complete transcript of the current conversation.

## Acceptance Criteria

- [ ] Exports the full message history including tool calls and results
- [ ] Output is in a readable format (markdown or JSON)

## Source Evidence

- `OpenHarness/src/openharness/commands/` — `/export` command
