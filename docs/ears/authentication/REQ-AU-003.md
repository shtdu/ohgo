# REQ-AU-003: Multi-Provider Backend Support

**Pattern:** Ubiquitous
**Capability:** Authentication

## Requirement

The system shall support multiple AI provider backends including Anthropic, OpenAI-compatible APIs, and subscription bridges (Claude CLI, Codex CLI).

## Acceptance Criteria

- [ ] Supports Anthropic API format
- [ ] Supports OpenAI-compatible API format
- [ ] Supports GitHub Copilot via OAuth
- [ ] Supports subscription bridges that proxy through existing CLI tools
- [ ] API format is selectable per provider profile

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `--api-format` flag
- `OpenHarness/src/openharness/bridge/` — subscription bridge implementations
