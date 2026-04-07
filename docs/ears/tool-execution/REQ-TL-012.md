# REQ-TL-012: Command Timeout Enforcement

**Pattern:** State-Driven
**Capability:** Tool Execution

## Requirement

While a shell command is executing, the system shall enforce the specified timeout and terminate the command if the duration is exceeded.

## Acceptance Criteria

- [ ] Default timeout applies when no explicit timeout is specified
- [ ] The system terminates the process and returns a timeout error message
- [ ] Partial output captured before timeout is included in the result

## Source Evidence

- `OpenHarness/src/openharness/tools/bash_tool.py` — `timeout_seconds` parameter
