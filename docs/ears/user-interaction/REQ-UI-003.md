# REQ-UI-003: Terminal User Interface

**Pattern:** Ubiquitous
**Capability:** User Interaction

## Requirement

The system shall render a terminal user interface that displays streaming AI responses, tool executions, and status indicators in real time.

## Acceptance Criteria

- [ ] Responses stream token-by-token to the terminal
- [ ] Tool invocations are displayed with parameters and results
- [ ] Progress indicators show during long-running operations
- [ ] The interface handles terminal resize events

## Source Evidence

- `OpenHarness/src/openharness/tui/` — React-based terminal UI
- `OpenHarness/frontend/` — frontend rendering components
