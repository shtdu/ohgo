# REQ-SM-004: Session Export

**Pattern:** Event-Driven
**Capability:** Session Management

## Requirement

When the user requests an export (`/export`), the system shall produce a complete transcript of the current conversation.

## Acceptance Criteria

- [ ] Exports the full message history including tool calls and results
- [ ] Output is in Markdown format with conversation turns as headings, or JSON with message-type discriminators
- [ ] The exported transcript preserves the chronological order of all messages
- [ ] When export fails (disk full, permission denied), the system reports the specific error and file path to the user

## Source Evidence

- `OpenHarness/src/openharness/commands/` — `/export` command
