# REQ-SM-005: Session Sharing

**Pattern:** Event-Driven
**Capability:** Session Management

## Requirement

When the user requests to share a session (`/share`), the system shall create a shareable artifact from the conversation transcript.

## Acceptance Criteria

- [ ] Produces a self-contained document including the full conversation with formatted tool results
- [ ] Includes the full conversation with formatted tool results
- [ ] The system provides a confirmation before creating the shareable artifact

## Source Evidence

- `OpenHarness/src/openharness/commands/` — `/share` command
