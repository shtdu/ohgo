# REQ-EX-002: Plugin Contribution Registration

**Pattern:** Event-Driven
**Capability:** Extensibility

## Requirement

When a plugin becomes active (first loaded per REQ-EX-001 or re-enabled per REQ-EX-008), the system shall register each of its contributions for use in the corresponding capability.

## Acceptance Criteria

- [ ] Plugin commands become available for invocation
- [ ] Plugin skills can be loaded on demand
- [ ] Plugin hooks execute at the appropriate execution points
- [ ] Plugin external tool servers are connected and their tools become available
- [ ] Contributions are registered when a plugin is first loaded (per REQ-EX-001) and when re-enabled (per REQ-EX-008)

## Source Evidence

- `OpenHarness/src/openharness/plugins/` — plugin contribution handling
