# REQ-EX-006: Hook Type Support

**Pattern:** Optional Feature
**Capability:** Extensibility

## Requirement

Where a hook of a supported type is configured, the system shall execute it according to its type's semantics: command execution, prompt evaluation, URL retrieval, or webhook notification.

## Acceptance Criteria

- [ ] Command-type hooks execute a configured action and return its output
- [ ] Prompt-type hooks produce an AI-generated response
- [ ] URL-type hooks retrieve content from a web address
- [ ] Webhook-type hooks send a notification to an external service

## Source Evidence

- `OpenHarness/src/openharness/hooks/` — hook type implementations
