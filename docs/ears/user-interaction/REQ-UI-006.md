# REQ-UI-006: Interactive User Prompts

**Pattern:** Event-Driven
**Capability:** User Interaction

## Requirement

When the agent needs user input for decisions, confirmations, or selections, the system shall present interactive prompts with selectable options and free-text input.

## Acceptance Criteria

- [ ] The system can present multiple-choice questions with 2-4 options
- [ ] The system supports free-text input when the user selects "Other"
- [ ] Prompts are used for user decisions, ambiguity resolution, and preference selection
- [ ] Tool execution approval requests are presented through the same prompt mechanism
- [ ] The prompt blocks agent execution until the user responds
- [ ] When the user does not respond within the timeout period, the system cancels the prompt and returns a timeout result to the agent
- [ ] When a prompt is cancelled (user interrupt or timeout), the agent receives a cancellation result and continues the session

## Source Evidence

- `OpenHarness/src/openharness/tools/ask_user_question_tool.py` — AskUserQuestionTool
