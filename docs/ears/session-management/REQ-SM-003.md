# REQ-SM-003: Session Resume

**Pattern:** Event-Driven
**Capability:** Session Management

## Requirement

When the user provides a session ID (`-r` flag), the system shall load the specified historical session.

## Acceptance Criteria

- [ ] Accepts a session ID as input
- [ ] Restores the full conversation state for that session
- [ ] Produces an error if the session ID does not exist

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `--resume` / `-r` flag
