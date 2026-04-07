# REQ-TL-008: Language Server Protocol Integration

**Pattern:** Optional Feature
**Capability:** Tool Execution

## Requirement

Where a language server is available for the file type, the system shall provide code intelligence operations including go-to-definition, find-references, hover, and document symbols.

## Acceptance Criteria

- [ ] Supports operations: goToDefinition, findReferences, hover, documentSymbol, workspaceSymbol
- [ ] Requires a running LSP server for the target language
- [ ] Returns structured results suitable for agent interpretation

## Source Evidence

- `OpenHarness/src/openharness/tools/lsp_tool.py`
