# Automation

Recurring and background task automation — scheduled jobs, cron expressions, and automatic prompt execution.

## Requirements

| ID | Title | Pattern |
|----|-------|---------|
| [REQ-AT-001](#req-at-001-background-task-execution) | Background Task Execution | Event-Driven |
| [REQ-AT-002](#req-at-002-background-task-lifecycle-management) | Background Task Lifecycle Management | State-Driven |
| [REQ-AT-003](#req-at-003-cron-scheduling) | Cron Scheduling | Optional Feature |
| [REQ-AT-004](#req-at-004-task-output-retrieval) | Task Output Retrieval | Event-Driven |
| [REQ-AT-005](#req-at-005-task-progress-tracking) | Task Progress Tracking | Event-Driven |

## Details

## REQ-AT-001: Background Task Execution

**Pattern:** Event-Driven

### Requirement

When a background task is created, the system shall execute it independently of the main conversation.

### Acceptance Criteria

- [ ] Tasks can run external commands
- [ ] Tasks can run agent-driven prompts
- [ ] Tasks execute independently of the main conversation loop
- [ ] Task state is queryable at any time
- [ ] When task execution fails (command not found, timeout), the task state transitions to failed with an error message


---

## REQ-AT-002: Background Task Lifecycle Management

**Pattern:** State-Driven

### Requirement

While a background task exists, the system shall manage its complete lifecycle from creation through termination. Progress tracking is covered separately by REQ-AT-005.

### Acceptance Criteria

- [ ] Tasks are created with unique IDs
- [ ] Task state transitions follow: pending, running, completed, or failed
- [ ] Tasks can be stopped by user request
- [ ] When a task creation fails (duplicate ID, invalid state), the system returns a descriptive error identifying the cause


---

## REQ-AT-003: Cron Scheduling

**Pattern:** Optional Feature

### Requirement

Where cron jobs are configured, the system shall execute specified commands or agent prompts on the defined cron schedule.

### Acceptance Criteria

- [ ] Cron expressions define the schedule
- [ ] Jobs can be enabled and disabled individually
- [ ] Jobs can be invoked outside their schedule via an immediate trigger
- [ ] Each scheduled execution records its outcome (success or failure)
- [ ] When a cron expression is invalid, the system rejects it with a parse error message identifying the malformed portion
- [ ] When cron job execution fails, the system logs the error and does not remove or disable the job


---

## REQ-AT-004: Task Output Retrieval

**Pattern:** Event-Driven

### Requirement

When the user requests task output, the system shall return the accumulated output up to a configurable size limit.

### Acceptance Criteria

- [ ] Returns output from completed or running tasks
- [ ] Respects a configurable maximum byte limit
- [ ] When output exceeds the size limit, the returned content is truncated and includes a note indicating truncation and the original size
- [ ] When output is requested for an invalid or expired task ID, the system returns an error indicating the task was not found


---

## REQ-AT-005: Task Progress Tracking

**Pattern:** Event-Driven

### Requirement

When a task updates its progress, the system shall persist the progress metadata (percentage, status note) for status queries.

### Acceptance Criteria

- [ ] Progress is a percentage value (0-100)
- [ ] A status note describes current activity
- [ ] Progress metadata is persisted and available to task retrieval operations (REQ-AT-002, REQ-AT-004)
- [ ] When an invalid progress value is provided (negative, over 100, or non-numeric), the system rejects the update and returns a validation error
