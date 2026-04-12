# REQ-AC-003: Inter-Agent Messaging

**Pattern:** Event-Driven
**Capability:** Agent Coordination

## Requirement

When an agent sends a message to another agent, the system shall deliver the message to the target agent's mailbox.

## Acceptance Criteria

- [ ] Messages are addressed by task ID
- [ ] The target agent receives the message in its input stream
- [ ] The sending agent continues execution without waiting for the receiving agent to process the message
- [ ] When a message is sent to a non-existent or terminated agent, the system returns a delivery failure error to the sender

## Source Evidence

- `OpenHarness/src/openharness/tools/send_message_tool.py`
- `OpenHarness/src/openharness/swarm/mailbox.py`
