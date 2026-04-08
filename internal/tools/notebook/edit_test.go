package notebook

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tools"
)

func TestNotebookEditTool_NameAndSchema(t *testing.T) {
	tool := NotebookEditTool{}
	assert.Equal(t, "notebook_edit", tool.Name())
	assert.Contains(t, tool.Description(), "notebook")
	schema := tool.InputSchema()
	assert.Equal(t, "object", schema["type"])
	required := schema["required"].([]string)
	assert.Contains(t, required, "notebook_path")
	assert.Contains(t, required, "new_source")
}

func createTestNotebook(t *testing.T, cells []notebookCell) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.ipynb")
	nb := notebook{
		NbFormat:      4,
		NbFormatMinor: 5,
		Metadata:      map[string]any{},
		Cells:         cells,
	}
	data, err := json.Marshal(nb)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0644))
	return path
}

func loadNotebook(t *testing.T, path string) notebook {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var nb notebook
	require.NoError(t, json.Unmarshal(data, &nb))
	return nb
}

func TestNotebookEditTool_ReplaceCellSource(t *testing.T) {
	path := createTestNotebook(t, []notebookCell{
		{CellType: "code", Metadata: map[string]any{}, Source: "print('hello')"},
		{CellType: "markdown", Metadata: map[string]any{}, Source: "# Title"},
	})

	cellNum := 0
	tool := NotebookEditTool{}
	args, _ := json.Marshal(map[string]any{
		"notebook_path": path,
		"cell_number":   cellNum,
		"new_source":    "print('world')",
		"edit_mode":     "replace",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Notebook edited")

	nb := loadNotebook(t, path)
	assert.Equal(t, 2, len(nb.Cells))
	assert.Equal(t, "print('world')", nb.Cells[0].Source)
	assert.Equal(t, "# Title", nb.Cells[1].Source)
}

func TestNotebookEditTool_InsertNewCell(t *testing.T) {
	path := createTestNotebook(t, []notebookCell{
		{CellType: "code", Metadata: map[string]any{}, Source: "print('first')"},
	})

	cellNum := 1
	tool := NotebookEditTool{}
	args, _ := json.Marshal(map[string]any{
		"notebook_path": path,
		"cell_number":   cellNum,
		"new_source":    "# New cell",
		"cell_type":     "markdown",
		"edit_mode":     "insert",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	nb := loadNotebook(t, path)
	assert.Equal(t, 2, len(nb.Cells))
	assert.Equal(t, "print('first')", nb.Cells[0].Source)
	assert.Equal(t, "# New cell", nb.Cells[1].Source)
	assert.Equal(t, "markdown", nb.Cells[1].CellType)
}

func TestNotebookEditTool_InsertCellAtBeginning(t *testing.T) {
	path := createTestNotebook(t, []notebookCell{
		{CellType: "code", Metadata: map[string]any{}, Source: "print('existing')"},
	})

	cellNum := 0
	tool := NotebookEditTool{}
	args, _ := json.Marshal(map[string]any{
		"notebook_path": path,
		"cell_number":   cellNum,
		"new_source":    "import os",
		"cell_type":     "code",
		"edit_mode":     "insert",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	nb := loadNotebook(t, path)
	assert.Equal(t, 2, len(nb.Cells))
	assert.Equal(t, "import os", nb.Cells[0].Source)
	assert.Equal(t, "print('existing')", nb.Cells[1].Source)
}

func TestNotebookEditTool_DeleteCell(t *testing.T) {
	path := createTestNotebook(t, []notebookCell{
		{CellType: "code", Metadata: map[string]any{}, Source: "first"},
		{CellType: "code", Metadata: map[string]any{}, Source: "second"},
		{CellType: "code", Metadata: map[string]any{}, Source: "third"},
	})

	cellNum := 1
	tool := NotebookEditTool{}
	args, _ := json.Marshal(map[string]any{
		"notebook_path": path,
		"cell_number":   cellNum,
		"new_source":    "",
		"edit_mode":     "delete",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	nb := loadNotebook(t, path)
	assert.Equal(t, 2, len(nb.Cells))
	assert.Equal(t, "first", nb.Cells[0].Source)
	assert.Equal(t, "third", nb.Cells[1].Source)
}

func TestNotebookEditTool_CreateNewNotebookOnInsert(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.ipynb")

	tool := NotebookEditTool{}
	args, _ := json.Marshal(map[string]any{
		"notebook_path": path,
		"new_source":    "print('hello')",
		"edit_mode":     "insert",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Notebook edited")

	nb := loadNotebook(t, path)
	assert.Equal(t, 4, nb.NbFormat)
	assert.Equal(t, 5, nb.NbFormatMinor)
	assert.Equal(t, 1, len(nb.Cells))
	assert.Equal(t, "print('hello')", nb.Cells[0].Source)
	assert.Equal(t, "code", nb.Cells[0].CellType)
}

func TestNotebookEditTool_CellNumberOutOfRange(t *testing.T) {
	path := createTestNotebook(t, []notebookCell{
		{CellType: "code", Metadata: map[string]any{}, Source: "only cell"},
	})

	cellNum := 5
	tool := NotebookEditTool{}
	args, _ := json.Marshal(map[string]any{
		"notebook_path": path,
		"cell_number":   cellNum,
		"new_source":    "replacement",
		"edit_mode":     "replace",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "out of range")
}

func TestNotebookEditTool_InvalidJSON(t *testing.T) {
	tool := NotebookEditTool{}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "invalid arguments")
}

func TestNotebookEditTool_MissingRequiredFields(t *testing.T) {
	tool := NotebookEditTool{}

	// Missing both notebook_path and new_source
	args, _ := json.Marshal(map[string]any{})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "notebook_path is required")
}

func TestNotebookEditTool_InvalidNotebookPath(t *testing.T) {
	tool := NotebookEditTool{}
	cellNum := 0
	args, _ := json.Marshal(map[string]any{
		"notebook_path": "/nonexistent/path/to/notebook.ipynb",
		"cell_number":   cellNum,
		"new_source":    "content",
		"edit_mode":     "replace",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "not found")
}

func TestNotebookEditTool_DefaultEditModeIsReplace(t *testing.T) {
	path := createTestNotebook(t, []notebookCell{
		{CellType: "code", Metadata: map[string]any{}, Source: "original"},
	})

	cellNum := 0
	tool := NotebookEditTool{}
	args, _ := json.Marshal(map[string]any{
		"notebook_path": path,
		"cell_number":   cellNum,
		"new_source":    "replaced",
		// no edit_mode — should default to "replace"
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	nb := loadNotebook(t, path)
	assert.Equal(t, "replaced", nb.Cells[0].Source)
}

func TestNotebookEditTool_InsertDefaultsToCodeCellType(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.ipynb")

	tool := NotebookEditTool{}
	args, _ := json.Marshal(map[string]any{
		"notebook_path": path,
		"new_source":    "x = 1",
		"edit_mode":     "insert",
		// no cell_type — should default to "code"
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	nb := loadNotebook(t, path)
	assert.Equal(t, "code", nb.Cells[0].CellType)
}

func TestNotebookEditTool_ReplaceRequiresCellNumber(t *testing.T) {
	path := createTestNotebook(t, []notebookCell{
		{CellType: "code", Metadata: map[string]any{}, Source: "data"},
	})

	tool := NotebookEditTool{}
	args, _ := json.Marshal(map[string]any{
		"notebook_path": path,
		"new_source":    "replacement",
		"edit_mode":     "replace",
		// no cell_number
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "cell_number is required")
}

func TestNotebookEditTool_DeleteRequiresCellNumber(t *testing.T) {
	path := createTestNotebook(t, []notebookCell{
		{CellType: "code", Metadata: map[string]any{}, Source: "data"},
	})

	tool := NotebookEditTool{}
	args, _ := json.Marshal(map[string]any{
		"notebook_path": path,
		"new_source":    "",
		"edit_mode":     "delete",
		// no cell_number
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "cell_number is required")
}

func TestNotebookEditTool_DeleteOutOfRange(t *testing.T) {
	path := createTestNotebook(t, []notebookCell{
		{CellType: "code", Metadata: map[string]any{}, Source: "only cell"},
	})

	cellNum := 10
	tool := NotebookEditTool{}
	args, _ := json.Marshal(map[string]any{
		"notebook_path": path,
		"cell_number":   cellNum,
		"new_source":    "",
		"edit_mode":     "delete",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content, "out of range")
}

func TestNotebookEditTool_WriteFormattedJSON(t *testing.T) {
	path := createTestNotebook(t, []notebookCell{
		{CellType: "code", Metadata: map[string]any{}, Source: "old"},
	})

	cellNum := 0
	tool := NotebookEditTool{}
	args, _ := json.Marshal(map[string]any{
		"notebook_path": path,
		"cell_number":   cellNum,
		"new_source":    "new",
		"edit_mode":     "replace",
	})
	result, err := tool.Execute(context.Background(), args)
	require.NoError(t, err)
	assert.False(t, result.IsError)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	// Should be formatted (indented) and end with newline
	assert.Contains(t, string(data), "  ")
	assert.Equal(t, "\n", string(data[len(data)-1:]))

	// Should round-trip cleanly
	var nb notebook
	require.NoError(t, json.Unmarshal(data, &nb))
}

func TestNotebookEditTool_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.ipynb")

	tool := NotebookEditTool{}
	args, _ := json.Marshal(map[string]any{
		"notebook_path": path,
		"new_source":    "data",
		"edit_mode":     "insert",
	})
	_, err := tool.Execute(ctx, args)
	assert.Error(t, err)
}

// Verify the tool satisfies the interface
var _ tools.Tool = NotebookEditTool{}
