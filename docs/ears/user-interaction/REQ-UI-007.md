# REQ-UI-007: TUI Themes

**Pattern:** Optional Feature
**Capability:** User Interaction

## Requirement

Where a theme is configured, the system shall apply the selected visual theme to the terminal interface, affecting colors, formatting, and layout.

## Acceptance Criteria

- [ ] The system provides at least the themes: default, dark, minimal, cyberpunk, solarized
- [ ] The theme is selectable via `--theme` flag or settings
- [ ] Theme changes take effect on the next rendered frame without requiring application restart
- [ ] When the specified theme configuration is invalid or cannot be loaded, the system falls back to the default theme

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `--theme` flag
- `OpenHarness/src/openharness/config/settings.py` — theme setting
