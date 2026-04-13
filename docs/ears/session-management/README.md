# Session Management

Conversation persistence, session resumption, conversation branching, and history management.

## Requirements

| ID | Title | Pattern |
|----|-------|---------|
| [REQ-SM-001](#req-sm-001-session-persistence) | Session Persistence | Ubiquitous |
| [REQ-SM-002](#req-sm-002-session-continue) | Session Continue | Event-Driven |
| [REQ-SM-003](#req-sm-003-session-resume) | Session Resume | Event-Driven |
| [REQ-SM-004](#req-sm-004-session-export) | Session Export | Event-Driven |
| [REQ-SM-005](#req-sm-005-session-sharing) | Session Sharing | Event-Driven |
| [REQ-SM-006](#req-sm-006-session-tagging) | Session Tagging | Event-Driven |
| [REQ-SM-007](#req-sm-007-session-rewind) | Session Rewind | Event-Driven |
| [REQ-SM-008](#req-sm-008-context-compaction) | Context Compaction | Event-Driven |

## Details

## REQ-SM-001: Session Persistence

**Pattern:** Ubiquitous

### Requirement

The system shall persist conversation state including message history and tool results so that sessions can be resumed after termination.

### Acceptance Criteria

- [ ] Session state is saved automatically at each turn completion
- [ ] Sessions are keyed by directory and session ID
- [ ] Session data survives process termination
- [ ] A persisted session can be restored to a state where the agent has access to the full conversation history
- [ ] When session storage is unavailable at startup, the system logs the error and operates in stateless mode for that session

### Source Evidence

- `OpenHarness/src/openharness/commands/` — `/resume`, `/continue` commands


---

## REQ-SM-002: Session Continue

**Pattern:** Event-Driven

### Requirement

When the user requests to continue a session (`-c` flag), the system shall load the most recent conversation for the current working directory.

### Acceptance Criteria

- [ ] Finds the session with the latest `updated_at` timestamp for the current directory
- [ ] Restores full message history
- [ ] The agent can reference and build upon information from the restored conversation history in subsequent responses
- [ ] When no previous session exists, the system starts a new session without error

### Source Evidence

- `OpenHarness/src/openharness/cli.py` — `--continue` / `-c` flag


---

## REQ-SM-003: Session Resume

**Pattern:** Event-Driven

### Requirement

When the user provides a session ID (`-r` flag), the system shall load the specified historical session.

### Acceptance Criteria

- [ ] Accepts a session ID as input
- [ ] Restores the full conversation state for that session
- [ ] Produces an error if the session ID does not exist

### Source Evidence

- `OpenHarness/src/openharness/cli.py` — `--resume` / `-r` flag


---

## REQ-SM-004: Session Export

**Pattern:** Event-Driven

### Requirement

When the user requests an export (`/export`), the system shall produce a complete transcript of the current conversation.

### Acceptance Criteria

- [ ] Exports the full message history including tool calls and results
- [ ] Output is in Markdown format with conversation turns as headings, or JSON with message-type discriminators
- [ ] The exported transcript preserves the chronological order of all messages
- [ ] When export fails (disk full, permission denied), the system reports the specific error and file path to the user

### Source Evidence

- `OpenHarness/src/openharness/commands/` — `/export` command


---

## REQ-SM-005: Session Sharing

**Pattern:** Event-Driven

### Requirement

When the user requests to share a session (`/share`), the system shall create a shareable artifact from the conversation transcript.

### Acceptance Criteria

- [ ] Produces a Markdown file containing the full conversation with a metadata header and formatted tool results
- [ ] Includes the full conversation with formatted tool results
- [ ] The system provides a confirmation before creating the shareable artifact
- [ ] When the share target file path is not writable, the system reports the specific error with the file path

### Source Evidence

- `OpenHarness/src/openharness/commands/` — `/share` command


---

## REQ-SM-006: Session Tagging

**Pattern:** Event-Driven

### Requirement

When the user tags a session (`/tag`), the system shall create a named snapshot of the current conversation state.

### Acceptance Criteria

- [ ] Accepts a tag name
- [ ] Creates a named checkpoint that can be referenced later
- [ ] Tagged sessions are listed in session history
- [ ] When the specified tag already exists, the system returns an error message without overwriting the existing tag

### Source Evidence

- `OpenHarness/src/openharness/commands/` — `/tag` command


---

## REQ-SM-007: Session Rewind

**Pattern:** Event-Driven

### Requirement

When the user rewinds a session (`/rewind`), the system shall remove the specified number of most recent conversation turns.

### Acceptance Criteria

- [ ] Accepts the number of turns to remove
- [ ] Removes both user and assistant messages for the specified turns
- [ ] The conversation continues from the rewound state
- [ ] When the rewind target is beyond the session history, the system reports an error and retains the current session state

### Source Evidence

- `OpenHarness/src/openharness/commands/` — `/rewind` command


---

## REQ-SM-008: Context Compaction

**Pattern:** Event-Driven

### Requirement

When the conversation reaches the compaction threshold (default: 90% of the model's context window capacity), the system shall compact older messages into a summary to free context window capacity.

### Acceptance Criteria

- [ ] Automatically triggers when token count reaches the compaction threshold (default: 90% of context window capacity)
- [ ] Preserves recent messages in full
- [ ] Compacted messages are replaced with a summary that is included in the conversation context
- [ ] The agent continues responding to new prompts using the compacted context
- [ ] When compaction fails to produce a usable summary, the system retains the original context and logs the failure

### Source Evidence

- `OpenHarness/src/openharness/commands/` — `/compact` command
- `OpenHarness/src/openharness/engine/` — auto-compaction logic
