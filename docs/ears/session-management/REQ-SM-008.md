# REQ-SM-008: Context Compaction

**Pattern:** Event-Driven
**Capability:** Session Management

## Requirement

When the conversation reaches the compaction threshold (default: 90% of the model's context window capacity), the system shall compact older messages while preserving key information to continue the session.

## Acceptance Criteria

- [ ] Automatically triggers when token count reaches the compaction threshold (default: 90% of context window capacity)
- [ ] Preserves recent messages in full
- [ ] Compacted messages are replaced with a summary that is included in the conversation context
- [ ] The agent continues responding to new prompts using the compacted context

## Source Evidence

- `OpenHarness/src/openharness/commands/` — `/compact` command
- Engine auto-compaction logic
