# REQ-PS-005: Tool Allow and Deny Lists

**Pattern:** Complex
**Capability:** Permissions and Safety

## Requirement

If a tool appears on the denied list, then the system shall block its execution regardless of permission mode; if a tool appears on the allowed list in default mode, then the system shall execute it without user confirmation.

## Acceptance Criteria

- [ ] Denied list takes precedence over all other settings
- [ ] When operating in default mode, the allowed list grants auto-approval; in other modes, the allowed list has no auto-approval effect
- [ ] Lists are configurable via CLI flags and settings
- [ ] Both built-in and MCP tools are subject to list filtering

## Source Evidence

- `OpenHarness/src/openharness/cli.py` — `--allowed-tools`, `--disallowed-tools`
- `OpenHarness/src/openharness/config/settings.py` — `permission.allowed_tools`, `permission.denied_tools`
