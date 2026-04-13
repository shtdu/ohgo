# Permissions & Safety

Permission modes, path-based access rules, tool-level allow/deny, and safety guardrails for destructive operations.

## Requirements

| ID | Title | Pattern |
|----|-------|---------|
| [REQ-PS-001](#req-ps-001-permission-mode-enforcement) | Permission Mode Enforcement | State-Driven |
| [REQ-PS-002](#req-ps-002-default-mode-user-confirmation) | Default Mode User Confirmation | Event-Driven |
| [REQ-PS-003](#req-ps-003-plan-mode-restrictions) | Plan Mode Restrictions | State-Driven |
| [REQ-PS-004](#req-ps-004-full-auto-mode-execution) | Full Auto Mode Execution | State-Driven |
| [REQ-PS-005](#req-ps-005-tool-allow-and-deny-lists) | Tool Allow and Deny Lists | Complex |
| [REQ-PS-006](#req-ps-006-path-permission-rules) | Path Permission Rules | Optional Feature |
| [REQ-PS-007](#req-ps-007-destructive-operation-warning) | Destructive Operation Warning | Event-Driven |
| [REQ-PS-008](#req-ps-008-permission-error-fail-safe) | Permission Error Fail-Safe | Unwanted Behaviour |

## Details

## REQ-PS-001: Permission Mode Enforcement

**Pattern:** State-Driven

### Requirement

While a permission mode is active, the system shall enforce tool execution permissions according to that mode's rules.

### Acceptance Criteria

- [ ] The system provides selectable permission modes (default, plan, and full_auto; details per REQ-PS-002, REQ-PS-003, REQ-PS-004)
- [ ] Every tool invocation is routed through the permission system before execution
- [ ] When the permission mode changes during a session, the new mode's rules apply to all subsequent tool invocations
- [ ] When an invalid permission mode is specified, the system defaults to the most restrictive mode and logs a warning


---

## REQ-PS-002: Default Mode User Confirmation

**Pattern:** Event-Driven

### Requirement

When the permission mode is default and the agent requests a sensitive tool execution, the system shall prompt the user for approval before proceeding.

### Acceptance Criteria

- [ ] Tools classified as write-capable (file write, file edit, command execution) require user approval
- [ ] Read-only tools bypass confirmation when that option is enabled in settings; otherwise all tools require confirmation
- [ ] The user can approve a single action, deny it, or approve all remaining actions for the session
- [ ] When the user denies a tool execution, the engine receives a rejection result and continues the session without error


---

## REQ-PS-003: Plan Mode Restrictions

**Pattern:** State-Driven

### Requirement

While the system is in plan mode, the system shall restrict tool execution to read-only operations and planning tools.

### Acceptance Criteria

- [ ] File write, edit, and bash tools are disabled
- [ ] File read, search, and planning tools remain available
- [ ] Any attempt to invoke a write-capable tool while in plan mode is rejected with an informative message


---

## REQ-PS-004: Full Auto Mode Execution

**Pattern:** State-Driven

### Requirement

While the system is in full auto mode, the system shall execute tools without user confirmation within configured boundaries.

### Acceptance Criteria

- [ ] Tools execute without user confirmation
- [ ] Denied tools list still blocks execution
- [ ] Path rules still restrict file operations
- [ ] Destructive operation warnings (per REQ-PS-007) remain active even in full auto mode


---

## REQ-PS-005: Tool Allow and Deny Lists

**Pattern:** Complex

### Requirement

If a tool appears on the denied list, then the system shall block its execution regardless of permission mode; if a tool appears on the allowed list in default mode, then the system shall execute it without user confirmation.

### Acceptance Criteria

- [ ] Denied list takes precedence over all other settings
- [ ] When operating in default mode, the allowed list grants auto-approval; in other modes, the allowed list has no auto-approval effect
- [ ] Lists are configurable via CLI flags and settings
- [ ] Both built-in and MCP tools are subject to list filtering


---

## REQ-PS-006: Path Permission Rules

**Pattern:** Optional Feature

### Requirement

Where path permission rules are configured, the system shall restrict file operations to the specified paths and block access to paths outside the rules.

### Acceptance Criteria

- [ ] Rules define allowed and denied path patterns
- [ ] File operations targeting paths outside the allowed set are rejected
- [ ] Path rules apply across all permission modes
- [ ] An access-denied message is returned identifying the blocked path
- [ ] When a path rule contains invalid syntax, the system rejects the rule at load time and reports the specific rule and error


---

## REQ-PS-007: Destructive Operation Warning

**Pattern:** Event-Driven

### Requirement

When the agent attempts a destructive operation (e.g., force push, file deletion, database drop), the system shall block execution pending explicit user confirmation before proceeding.

### Acceptance Criteria

- [ ] Destructive patterns are detected in tool inputs using pattern matching
- [ ] Presents a warning containing the matched destructive pattern name and the specific file or command path
- [ ] Execution remains blocked until the user explicitly confirms or denies the operation
- [ ] When the user confirms, execution proceeds with the destructive operation; when denied, execution is cancelled


---

## REQ-PS-008: Permission Error Fail-Safe

**Pattern:** Unwanted Behaviour

### Requirement

If the permission system encounters an error during evaluation, the system shall refuse tool execution and report the error to the user.

### Acceptance Criteria

- [ ] Tool execution is blocked when permission checking fails
- [ ] The user receives an error message containing the tool name, the permission rule that failed, and the failure reason
- [ ] The error is logged with the tool name, permission context, and timestamp
