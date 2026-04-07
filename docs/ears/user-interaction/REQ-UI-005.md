# REQ-UI-005: Channel Gateway

**Pattern:** Optional Feature
**Capability:** User Interaction

## Requirement

Where a channel gateway is configured, the system shall receive messages from external messaging platforms and respond within the originating conversation thread.

## Acceptance Criteria

- [ ] Supports Telegram, Slack, Discord, Feishu, DingTalk, WhatsApp, Matrix, QQ, and MoChat
- [ ] Messages route to the agent engine and responses return to the originating channel
- [ ] Each channel conversation maintains independent session context
- [ ] The gateway can run as a persistent background service

## Source Evidence

- `OpenHarness/src/openharness/channels/` — channel integrations
- `OpenHarness/ohmo/cli.py` — `gateway` subcommand
