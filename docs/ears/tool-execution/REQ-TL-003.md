# REQ-TL-003: Shell Command Execution

**Pattern:** Event-Driven
**Capability:** Tool Execution

## Requirement

When the agent invokes the bash tool, the system shall execute the specified shell command in the working directory and return captured stdout and stderr output.

## Acceptance Criteria

- [ ] Commands execute in the configured working directory
- [ ] The system captures both stdout and stderr
- [ ] A configurable timeout (1-600 seconds) terminates long-running commands
- [ ] The working directory persists between sequential command invocations

## Source Evidence

- `OpenHarness/src/openharness/tools/bash_tool.py`
