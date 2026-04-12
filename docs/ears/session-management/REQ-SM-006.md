# REQ-SM-006: Session Tagging

**Pattern:** Event-Driven
**Capability:** Session Management

## Requirement

When the user tags a session (`/tag`), the system shall create a named snapshot of the current conversation state.

## Acceptance Criteria

- [ ] Accepts a tag name
- [ ] Creates a named checkpoint that can be referenced later
- [ ] Tagged sessions are listed in session history
- [ ] When the specified tag already exists, the system returns an error message without overwriting the existing tag

## Source Evidence

- `OpenHarness/src/openharness/commands/` — `/tag` command
