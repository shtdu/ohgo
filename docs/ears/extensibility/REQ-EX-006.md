# REQ-EX-006: Hook Type Support

**Pattern:** Ubiquitous
**Capability:** Extensibility

## Requirement

The system shall support multiple hook types for lifecycle customization: command hooks (shell), prompt hooks (LLM), URL hooks (fetch), and webhook hooks (HTTP).

## Acceptance Criteria

- [ ] Command hooks execute shell commands and capture output
- [ ] Prompt hooks invoke the LLM with a custom prompt
- [ ] URL hooks fetch content from a URL
- [ ] Webhook hooks send HTTP requests to external services

## Source Evidence

- `OpenHarness/src/openharness/hooks/` — hook type implementations
