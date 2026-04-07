# REQ-EX-004: On-Demand Skill Loading

**Pattern:** Event-Driven
**Capability:** Extensibility

## Requirement

When the agent invokes a skill, the system shall load the skill's markdown content and inject it into the agent's context for the current turn.

## Acceptance Criteria

- [ ] Skills are loaded from bundled, user, and plugin sources
- [ ] Each skill provides a name and description in its metadata
- [ ] The skill content becomes part of the agent's instructions for execution
- [ ] Skills are loaded on demand, not all at startup

## Source Evidence

- `OpenHarness/src/openharness/skills/` — skill registry and loading
- `OpenHarness/src/openharness/tools/skill_tool.py`
