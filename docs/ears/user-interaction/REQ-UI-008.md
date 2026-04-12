# REQ-UI-008: Vim Input Mode

**Pattern:** Optional Feature
**Capability:** User Interaction

## Requirement

Where vim mode is enabled, the system shall provide vim-style keybindings for input field navigation and editing.

## Acceptance Criteria

- [ ] Vim mode is toggleable via `/vim` command or settings
- [ ] Supports modal editing (normal mode, insert mode)
- [ ] Supports h/j/k/l for movement, i/a for insert mode entry, Esc for normal mode, and w/b for word navigation
- [ ] When vim mode configuration cannot be loaded, the system falls back to default (non-vim) input handling

## Source Evidence

- `OpenHarness/src/openharness/commands/` — `/vim` slash command
- Settings: `vim_mode` key
