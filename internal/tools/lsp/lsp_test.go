package lsp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tools"
)

func TestLspTool_Name(t *testing.T) {
	assert.Equal(t, "lsp", LspTool{}.Name())
}

func TestLspTool_InvalidOperation(t *testing.T) {
	tool := LspTool{}
	args, _ := json.Marshal(map[string]string{"operation": "invalid_op"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid operation")
}

func TestLspTool_ValidOperations(t *testing.T) {
	ops := []string{"document_symbol", "workspace_symbol", "go_to_definition", "find_references", "hover"}
	filePath := "test.go"
	line := 1
	char := 1
	query := "main"

	for _, op := range ops {
		t.Run(op, func(t *testing.T) {
			input := map[string]any{"operation": op}
			switch op {
			case "document_symbol":
				input["file_path"] = filePath
			case "go_to_definition", "find_references", "hover":
				input["file_path"] = filePath
				input["line"] = line
				input["character"] = char
			case "workspace_symbol":
				input["query"] = query
			}

			args, _ := json.Marshal(input)
			result, err := LspTool{}.Execute(context.Background(), args)
			require.NoError(t, err)
			assert.False(t, result.IsError, op)
			assert.Contains(t, result.Content, op)
		})
	}
}

func TestLspTool_MissingFilePath(t *testing.T) {
	tool := LspTool{}
	args, _ := json.Marshal(map[string]string{"operation": "document_symbol"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "file_path is required")
}

func TestLspTool_MissingPosition(t *testing.T) {
	tool := LspTool{}
	args, _ := json.Marshal(map[string]any{"operation": "go_to_definition", "file_path": "test.go"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "line and character are required")
}

func TestLspTool_MissingQuery(t *testing.T) {
	tool := LspTool{}
	args, _ := json.Marshal(map[string]string{"operation": "workspace_symbol"})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "query is required")
}

func TestLspTool_InvalidJSON(t *testing.T) {
	tool := LspTool{}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{bad`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

var _ tools.Tool = LspTool{}
