# REQ-PS-006: Path Permission Rules

**Pattern:** Optional Feature
**Capability:** Permissions and Safety

## Requirement

Where path permission rules are configured, the system shall restrict file operations to the specified paths and block access to paths outside the rules.

## Acceptance Criteria

- [ ] Rules define allowed and denied path patterns
- [ ] File operations targeting paths outside the allowed set are rejected
- [ ] Path rules apply across all permission modes
- [ ] An access-denied message is returned identifying the blocked path

## Source Evidence

- `OpenHarness/src/openharness/settings.py` — `permission.path_rules`
