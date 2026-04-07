# REQ-SM-008: Context Compaction

**Pattern:** Event-Driven
**Capability:** Session Management

## Requirement

When the conversation approaches context window limits, the system shall compact older messages while preserving key information to continue the session.

## Acceptance Criteria

- [ ] Automatically triggers when context length approaches the model limit
- [ ] Preserves recent messages in full
- [ ] Summarizes older messages to retain key information
- [ ] The agent continues functioning after compaction without user intervention

## Source Evidence

- `OpenHarness/src/openharness/commands/` — `/compact` command
- Engine auto-compaction logic
