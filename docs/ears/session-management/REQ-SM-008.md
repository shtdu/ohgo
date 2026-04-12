# REQ-SM-008: Context Compaction

**Pattern:** Event-Driven
**Capability:** Session Management

## Requirement

When the conversation reaches the compaction threshold (default: 90% of the model's context window capacity), the system shall compact older messages into a summary to free context window capacity.

## Acceptance Criteria

- [ ] Automatically triggers when token count reaches the compaction threshold (default: 90% of context window capacity)
- [ ] Preserves recent messages in full
- [ ] Compacted messages are replaced with a summary that is included in the conversation context
- [ ] The agent continues responding to new prompts using the compacted context
- [ ] When compaction fails to produce a usable summary, the system retains the original context and logs the failure

## Source Evidence

- `OpenHarness/src/openharness/commands/` — `/compact` command
- `OpenHarness/src/openharness/engine/` — auto-compaction logic
