# REQ-SM-007: Session Rewind

**Pattern:** Event-Driven
**Capability:** Session Management

## Requirement

When the user rewinds a session (`/rewind`), the system shall remove the specified number of most recent conversation turns.

## Acceptance Criteria

- [ ] Accepts the number of turns to remove
- [ ] Removes both user and assistant messages for the specified turns
- [ ] The conversation continues from the rewound state

## Source Evidence

- `OpenHarness/src/openharness/commands/` — `/rewind` command
