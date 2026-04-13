# User Interaction

How users communicate with the system — CLI, TUI, slash commands, channel gateways, and interactive prompts.

## Requirements

| ID | Title | Pattern |
|----|-------|---------|
| [REQ-UI-001](#req-ui-001-command-line-interface) | Command-Line Interface | Ubiquitous |
| [REQ-UI-002](#req-ui-002-cli-flags-and-options) | CLI Flags and Options | Ubiquitous |
| [REQ-UI-003](#req-ui-003-terminal-user-interface) | Terminal User Interface | Ubiquitous |
| [REQ-UI-004](#req-ui-004-slash-commands) | Slash Commands | Event-Driven |
| [REQ-UI-005](#req-ui-005-channel-gateway) | Channel Gateway | Optional Feature |
| [REQ-UI-006](#req-ui-006-interactive-user-prompts) | Interactive User Prompts | Event-Driven |
| [REQ-UI-007](#req-ui-007-tui-themes) | TUI Themes | Optional Feature |
| [REQ-UI-008](#req-ui-008-vim-input-mode) | Vim Input Mode | Optional Feature |

## Details

## REQ-UI-001: Command-Line Interface

**Pattern:** Ubiquitous

### Requirement

The system shall provide a command-line interface that accepts natural language prompts as the primary interaction method.

### Acceptance Criteria

- [ ] The system provides a `og` command that launches the interface
- [ ] The system accepts free-text prompts as positional arguments
- [ ] The system supports interactive mode when launched without a prompt
- [ ] The system returns a non-zero exit code on failure
- [ ] When the model service is unreachable at startup, the system reports a connection error before entering interactive mode


---

## REQ-UI-002: CLI Flags and Options

**Pattern:** Ubiquitous

### Requirement

The system shall accept model selection, permission mode, effort level, and output format options via command-line flags that override default settings for the session.

### Acceptance Criteria

- [ ] `--model` / `-m` selects the AI model by alias or full ID
- [ ] `--permission-mode` sets the permission mode (default, plan, full_auto)
- [ ] `--effort` sets the reasoning effort level
- [ ] `--output-format` sets output format (text, json, stream-json)
- [ ] `--print` / `-p` prints response and exits (non-interactive)
- [ ] `--max-turns` limits agentic turns
- [ ] CLI flags override values from the settings file for the duration of the session
- [ ] When an invalid flag value is provided (unknown model, unsupported permission mode), the system reports the error and exits with a non-zero status code


---

## REQ-UI-003: Terminal User Interface

**Pattern:** Ubiquitous

### Requirement

The system shall render a terminal user interface that displays streaming AI responses, tool executions, and status indicators in real time.

### Acceptance Criteria

- [ ] Responses stream token-by-token to the terminal
- [ ] Tool invocations are displayed with parameters and results
- [ ] Progress indicators show during tool execution and API streaming
- [ ] The interface handles terminal resize events
- [ ] When the terminal is too small to render the interface, the system displays a minimum-size warning message


---

## REQ-UI-004: Slash Commands

**Pattern:** Event-Driven

### Requirement

When a user enters a slash command (e.g., `/help`, `/commit`, `/plan`), the system shall execute the corresponding built-in or plugin-registered command.

### Acceptance Criteria

- [ ] The system recognizes commands starting with `/`
- [ ] Built-in commands include at minimum: help, exit, clear, commit, plan, status, config
- [ ] Plugins can register additional slash commands
- [ ] Unknown commands produce a descriptive error message
- [ ] When a built-in command fails during execution, the system reports the error and returns to the prompt loop
- [ ] When a plugin registers a command name matching a built-in command, the plugin command is namespaced and does not override the built-in


---

## REQ-UI-005: Channel Gateway

**Pattern:** Optional Feature

### Requirement

Where a channel gateway is configured, the system shall receive messages from external messaging platforms and respond within the originating conversation thread.

### Acceptance Criteria

- [ ] Supports Telegram, Slack, Discord, Feishu, DingTalk, WhatsApp, Matrix, QQ, and MoChat
- [ ] Messages route to the agent engine and responses return to the originating channel
- [ ] Each channel conversation maintains independent session context
- [ ] The gateway can run as a persistent background service
- [ ] When a gateway connection times out, the system retries up to 3 times with exponential backoff (maximum delay 30 seconds) and reports failure after exhausting retries
- [ ] When channel authentication fails (invalid token, expired credentials), the system logs the error and does not attempt to process messages for that channel


---

## REQ-UI-006: Interactive User Prompts

**Pattern:** Event-Driven

### Requirement

When the agent needs user input for decisions, confirmations, or selections, the system shall present interactive prompts with selectable options and free-text input.

### Acceptance Criteria

- [ ] The system can present multiple-choice questions with 2-4 options
- [ ] The system supports free-text input when the user selects "Other"
- [ ] Prompts are used for user decisions, ambiguity resolution, and preference selection
- [ ] Tool execution approval requests are presented through the same prompt mechanism
- [ ] The prompt blocks agent execution until the user responds
- [ ] When the user does not respond within the timeout period, the system cancels the prompt and returns a timeout result to the agent
- [ ] When a prompt is cancelled (user interrupt or timeout), the agent receives a cancellation result and continues the session


---

## REQ-UI-007: TUI Themes

**Pattern:** Optional Feature

### Requirement

Where a theme is configured, the system shall apply the selected visual theme to the terminal interface, affecting colors, formatting, and layout.

### Acceptance Criteria

- [ ] The system provides at least the themes: default, dark, minimal, cyberpunk, solarized
- [ ] The theme is selectable via `--theme` flag or settings
- [ ] Theme changes take effect on the next rendered frame without requiring application restart
- [ ] When the specified theme configuration is invalid or cannot be loaded, the system falls back to the default theme


---

## REQ-UI-008: Vim Input Mode

**Pattern:** Optional Feature

### Requirement

Where vim mode is enabled, the system shall provide vim-style keybindings for input field navigation and editing.

### Acceptance Criteria

- [ ] Vim mode is toggleable via `/vim` command or settings
- [ ] Supports modal editing (normal mode, insert mode)
- [ ] Supports h/j/k/l for movement, i/a for insert mode entry, Esc for normal mode, and w/b for word navigation
- [ ] When vim mode configuration cannot be loaded, the system falls back to default (non-vim) input handling
