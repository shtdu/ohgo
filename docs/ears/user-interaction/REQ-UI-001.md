# REQ-UI-001: Command-Line Interface

**Pattern:** Ubiquitous
**Capability:** User Interaction

## Requirement

The system shall provide a command-line interface that accepts natural language prompts as the primary interaction method.

## Acceptance Criteria

- [ ] The system provides a `og` command that launches the interface
- [ ] The system accepts free-text prompts as positional arguments
- [ ] The system supports interactive mode when launched without a prompt
- [ ] The system returns a non-zero exit code on failure
- [ ] When the model service is unreachable at startup, the system reports a connection error before entering interactive mode

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — CLI entry point with prompt handling
- `OpenHarness/src/openharness/__main__.py` — module invocation
