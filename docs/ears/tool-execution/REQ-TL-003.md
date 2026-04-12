# REQ-TL-003: Shell Command Execution

**Pattern:** Event-Driven
**Capability:** Tool Execution

## Requirement

When the agent invokes the command execution tool, the system shall execute the specified command in the configured working directory and return captured output.

## Acceptance Criteria

- [ ] Commands execute in the configured working directory
- [ ] Captures both standard output and standard error output
- [ ] Partial output captured before timeout is included in the result
- [ ] A default timeout applies when no explicit timeout is specified; exceeding it terminates the command and returns a timeout error message
- [ ] The working directory persists between sequential command invocations
- [ ] When a shell command returns a non-zero exit code, the tool returns both stdout and stderr to the agent
- [ ] When the command executable is not found or execution is denied by permissions, the tool returns an error with the command name and failure reason

## Source Evidence

- `OpenHarness/src/openharness/tools/bash_tool.py`
