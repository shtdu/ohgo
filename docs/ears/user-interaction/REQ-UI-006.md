# REQ-UI-006: Interactive User Prompts

**Pattern:** Event-Driven
**Capability:** User Interaction

## Requirement

When the agent needs user input for decisions, confirmations, or selections, the system shall present interactive prompts with selectable options and free-text input.

## Acceptance Criteria

- [ ] The system can present multiple-choice questions with 2-4 options
- [ ] The system supports free-text input when the user selects "Other"
- [ ] Prompts are used for tool approval, ambiguity resolution, and user preferences
- [ ] The prompt blocks agent execution until the user responds

## Source Evidence

- `OpenHarness/src/openharness/tools/ask_user_question_tool.py` — AskUserQuestionTool
