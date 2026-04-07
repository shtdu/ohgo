# REQ-PS-006: Path Permission Rules

**Pattern:** Complex
**Capability:** Permissions and Safety

## Requirement

If path permission rules are configured, then the system shall restrict file operations to the specified paths and block access to paths outside the rules.

## Acceptance Criteria

- [ ] Rules define allowed and denied path patterns
- [ ] File read, write, and edit tools check paths against rules
- [ ] Path rules apply across all permission modes

## Source Evidence

- `OpenHarness/src/openharness/settings.py` — `permission.path_rules`
