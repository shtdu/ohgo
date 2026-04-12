# REQ-UI-003: Terminal User Interface

**Pattern:** Ubiquitous
**Capability:** User Interaction

## Requirement

The system shall render a terminal user interface that displays streaming AI responses, tool executions, and status indicators in real time.

## Acceptance Criteria

- [ ] Responses stream token-by-token to the terminal
- [ ] Tool invocations are displayed with parameters and results
- [ ] Progress indicators show during tool execution and API streaming
- [ ] The interface handles terminal resize events
- [ ] When the terminal is too small to render the interface, the system displays a minimum-size warning message

## Source Evidence

- `OpenHarness/src/openharness/ui/` — terminal UI components
- `OpenHarness/frontend/` — frontend rendering components
