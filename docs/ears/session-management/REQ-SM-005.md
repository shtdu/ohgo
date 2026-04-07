# REQ-SM-005: Session Sharing

**Pattern:** Event-Driven
**Capability:** Session Management

## Requirement

When the user requests to share a session (`/share`), the system shall create a shareable artifact from the conversation transcript.

## Acceptance Criteria

- [ ] Produces a self-contained shareable document
- [ ] Includes the full conversation with formatted tool results

## Source Evidence

- `OpenHarness/src/openharness/commands/` — `/share` command
