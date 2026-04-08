// Package lsp implements the LSP tool for code intelligence operations.
// This is a stub implementation that validates inputs but does not connect
// to language servers. Full LSP integration will be added in a future phase.
package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shtdu/ohgo/internal/tools"
)

var validOperations = map[string]bool{
	"document_symbol":  true,
	"workspace_symbol": true,
	"go_to_definition": true,
	"find_references":  true,
	"hover":            true,
}

type lspInput struct {
	Operation string  `json:"operation"`
	FilePath  *string `json:"file_path"`
	Symbol    *string `json:"symbol"`
	Line      *int    `json:"line"`
	Character *int    `json:"character"`
	Query     *string `json:"query"`
}

// LspTool provides code intelligence via LSP operations.
type LspTool struct{}

func (LspTool) Name() string { return "lsp" }

func (LspTool) Description() string {
	return "Code intelligence operations: go_to_definition, find_references, hover, document_symbol, workspace_symbol. Requires an LSP server."
}

func (LspTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"operation": map[string]any{
				"type":        "string",
				"enum":        []string{"document_symbol", "workspace_symbol", "go_to_definition", "find_references", "hover"},
				"description": "The code intelligence operation to perform",
			},
			"file_path": map[string]any{
				"type":        "string",
				"description": "Path to the source file for file-based operations",
			},
			"line": map[string]any{
				"type":        "integer",
				"description": "1-based line number",
				"minimum":     1,
			},
			"character": map[string]any{
				"type":        "integer",
				"description": "1-based character offset",
				"minimum":     1,
			},
			"query": map[string]any{
				"type":        "string",
				"description": "Query for workspace_symbol search",
			},
		},
		"required":             []string{"operation"},
		"additionalProperties": false,
	}
}

func (LspTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	var input lspInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	// Validate operation
	if !validOperations[input.Operation] {
		valid := make([]string, 0, len(validOperations))
		for op := range validOperations {
			valid = append(valid, op)
		}
		return tools.Result{
			Content: fmt.Sprintf("invalid operation %q; valid: %s", input.Operation, strings.Join(valid, ", ")),
			IsError: true,
		}, nil
	}

	// Validate required fields per operation
	switch input.Operation {
	case "document_symbol":
		if input.FilePath == nil || *input.FilePath == "" {
			return tools.Result{
				Content: "file_path is required for document_symbol operation",
				IsError: true,
			}, nil
		}
	case "go_to_definition", "find_references", "hover":
		if input.FilePath == nil || *input.FilePath == "" {
			return tools.Result{
				Content: fmt.Sprintf("file_path is required for %s operation", input.Operation),
				IsError: true,
			}, nil
		}
		if input.Line == nil || input.Character == nil {
			return tools.Result{
				Content: fmt.Sprintf("line and character are required for %s operation", input.Operation),
				IsError: true,
			}, nil
		}
	case "workspace_symbol":
		if input.Query == nil || *input.Query == "" {
			return tools.Result{
				Content: "query is required for workspace_symbol operation",
				IsError: true,
			}, nil
		}
	}

	// Stub response
	return tools.Result{
		Content: fmt.Sprintf("LSP %s: not connected to a language server. Configure an LSP server to enable code intelligence.", input.Operation),
	}, nil
}
