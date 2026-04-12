# REQ-SM-005: Session Sharing

**Pattern:** Event-Driven
**Capability:** Session Management

## Requirement

When the user requests to share a session (`/share`), the system shall create a shareable artifact from the conversation transcript.

## Acceptance Criteria

- [ ] Produces a Markdown file containing the full conversation with a metadata header and formatted tool results
- [ ] Includes the full conversation with formatted tool results
- [ ] The system provides a confirmation before creating the shareable artifact
- [ ] When the share target file path is not writable, the system reports the specific error with the file path

## Source Evidence

- `OpenHarness/src/openharness/commands/` — `/share` command
