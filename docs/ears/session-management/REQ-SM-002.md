# REQ-SM-002: Session Continue

**Pattern:** Event-Driven
**Capability:** Session Management

## Requirement

When the user requests to continue a session (`-c` flag), the system shall load the most recent conversation for the current working directory.

## Acceptance Criteria

- [ ] Finds the most recent session for the current directory
- [ ] Restores full message history
- [ ] The agent can reference and build upon information from the restored conversation history in subsequent responses

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `--continue` / `-c` flag
