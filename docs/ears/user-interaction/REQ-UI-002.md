# REQ-UI-002: CLI Flags and Options

**Pattern:** Ubiquitous
**Capability:** User Interaction

## Requirement

The system shall accept model selection, permission mode, effort level, and output format options via command-line flags that override default settings for the session.

## Acceptance Criteria

- [ ] `--model` / `-m` selects the AI model by alias or full ID
- [ ] `--permission-mode` sets the permission mode (default, plan, full_auto)
- [ ] `--effort` sets the reasoning effort level
- [ ] `--output-format` sets output format (text, json, stream-json)
- [ ] `--print` / `-p` prints response and exits (non-interactive)
- [ ] `--max-turns` limits agentic turns
- [ ] CLI flags override values from the settings file for the duration of the session

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — argparse definitions for all flags
