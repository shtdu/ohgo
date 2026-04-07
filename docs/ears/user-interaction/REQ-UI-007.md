# REQ-UI-007: TUI Themes

**Pattern:** Optional Feature
**Capability:** User Interaction

## Requirement

Where a theme is configured, the system shall apply the selected visual theme to the terminal interface, affecting colors, formatting, and layout.

## Acceptance Criteria

- [ ] The system provides at least the themes: default, dark, minimal, cyberpunk, solarized
- [ ] The theme is selectable via `--theme` flag or settings
- [ ] Theme changes apply immediately without restart

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `--theme` flag
- `OpenHarness/src/openharness/settings.py` — theme setting
