# REQ-SM-001: Session Persistence

**Pattern:** Ubiquitous
**Capability:** Session Management

## Requirement

The system shall persist conversation state including message history and tool results so that sessions can be resumed after termination.

## Acceptance Criteria

- [ ] Session state is saved automatically during conversation
- [ ] Sessions are keyed by directory and session ID
- [ ] Session data survives process termination
- [ ] A persisted session can be restored to a state where the agent has access to the full conversation history

## Source Evidence

- `OpenHarness/src/openharness/commands/` — `/resume`, `/continue` commands
