# REQ-TL-008: Code Intelligence

**Pattern:** Optional Feature
**Capability:** Tool Execution

## Requirement

Where a language intelligence service is available for the file type, the system shall provide code intelligence and navigation operations.

## Acceptance Criteria

- [ ] Supports operations: goToDefinition, findReferences, hover, documentSymbol, workspaceSymbol
- [ ] Results include the source file path and line number for each symbol
- [ ] Returns structured results suitable for agent interpretation
- [ ] When no language intelligence service is available for the target file type, the system returns a descriptive error

## Source Evidence

- `OpenHarness/src/openharness/tools/lsp_tool.py`
