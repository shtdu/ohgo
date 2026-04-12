# REQ-EX-006: Hook Type Support

**Pattern:** Optional Feature
**Capability:** Extensibility

## Requirement

Where a hook of a supported type is configured, the system shall execute it according to the deserialization and validation rules defined for its declared parameter type.

## Acceptance Criteria

- [ ] Command-type hooks execute a configured action and return its output
- [ ] Prompt-type hooks produce an AI-generated response
- [ ] URL-type hooks retrieve content from a web address
- [ ] Webhook-type hooks send a notification to an external service
- [ ] When a hook execution fails (command not found, network timeout), the system logs the error and continues the session without the hook result

## Source Evidence

- `OpenHarness/src/openharness/hooks/` — hook type implementations
