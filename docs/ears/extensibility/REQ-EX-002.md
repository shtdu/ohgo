# REQ-EX-002: Plugin Contribution Registration

**Pattern:** Complex
**Capability:** Extensibility

## Requirement

If a plugin provides commands, skills, hooks, or MCP servers, then the system shall register each contribution in the appropriate subsystem during plugin loading.

## Acceptance Criteria

- [ ] Plugin commands are registered as slash commands
- [ ] Plugin skills are added to the skill registry
- [ ] Plugin hooks are registered for lifecycle events
- [ ] Plugin MCP servers are connected and tools registered

## Source Evidence

- `OpenHarness/src/openharness/plugins/` — plugin contribution handling
