// Package notebook implements the notebook_edit tool for editing Jupyter notebook files.
package notebook

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/shtdu/ohgo/internal/tools"
)

// notebook represents the top-level structure of a .ipynb file.
type notebook struct {
	NbFormat      int            `json:"nbformat"`
	NbFormatMinor int            `json:"nbformat_minor"`
	Metadata      map[string]any `json:"metadata"`
	Cells         []notebookCell `json:"cells"`
}

// notebookCell represents a single cell in a Jupyter notebook.
type notebookCell struct {
	CellType string         `json:"cell_type"`
	ID       string         `json:"id,omitempty"`
	Metadata map[string]any `json:"metadata"`
	Source   any            `json:"source"`
}

// notebookEditInput holds the parsed arguments for the notebook_edit tool.
type notebookEditInput struct {
	NotebookPath string `json:"notebook_path"`
	CellID       string `json:"cell_id"`
	CellNumber   *int   `json:"cell_number"`
	NewSource    string `json:"new_source"`
	CellType     string `json:"cell_type"`
	EditMode     string `json:"edit_mode"`
}

// NotebookEditTool edits cells in Jupyter notebook (.ipynb) files.
type NotebookEditTool struct{}

func (NotebookEditTool) Name() string { return "notebook_edit" }

func (NotebookEditTool) Description() string {
	return "Completely replaces the contents of a specific cell in a Jupyter notebook (.ipynb file) with new source."
}

func (NotebookEditTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"notebook_path": map[string]any{
				"type":        "string",
				"description": "Absolute path to the Jupyter notebook file (.ipynb)",
			},
			"cell_id": map[string]any{
				"type":        "string",
				"description": "ID of the cell to edit. Use with cell_number for cell selection.",
			},
			"cell_number": map[string]any{
				"type":        "integer",
				"description": "0-indexed cell number",
				"minimum":     0,
			},
			"new_source": map[string]any{
				"type":        "string",
				"description": "New source for the cell",
			},
			"cell_type": map[string]any{
				"type":        "string",
				"enum":        []string{"code", "markdown"},
				"description": "Cell type for insert mode",
			},
			"edit_mode": map[string]any{
				"type":        "string",
				"enum":        []string{"replace", "insert", "delete"},
				"default":     "replace",
				"description": "Edit mode: replace cell, insert new cell, or delete cell",
			},
		},
		"required":             []string{"notebook_path", "new_source"},
		"additionalProperties": false,
	}
}

func (NotebookEditTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	var input notebookEditInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	// Validate required fields
	if input.NotebookPath == "" {
		return tools.Result{Content: "notebook_path is required", IsError: true}, nil
	}

	// Check context
	select {
	case <-ctx.Done():
		return tools.Result{}, ctx.Err()
	default:
	}

	// Resolve path
	path := tools.ResolvePath(input.NotebookPath)

	// Determine edit mode, default to "replace"
	editMode := input.EditMode
	if editMode == "" {
		editMode = "replace"
	}

	// Load or create notebook
	var nb notebook
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && editMode == "insert" {
			nb = newNotebook()
		} else if os.IsNotExist(err) {
			return tools.Result{Content: fmt.Sprintf("Notebook file not found: %s", input.NotebookPath), IsError: true}, nil
		} else {
			return tools.Result{Content: fmt.Sprintf("Cannot read notebook: %v", err), IsError: true}, nil
		}
	} else {
		if err := json.Unmarshal(data, &nb); err != nil {
			return tools.Result{Content: fmt.Sprintf("Invalid notebook JSON: %v", err), IsError: true}, nil
		}
	}

	// Apply the edit
	switch editMode {
	case "replace":
		if err := replaceCell(&nb, input); err != nil {
			return tools.Result{Content: err.Error(), IsError: true}, nil
		}
	case "insert":
		insertCell(&nb, input)
	case "delete":
		if err := deleteCell(&nb, input); err != nil {
			return tools.Result{Content: err.Error(), IsError: true}, nil
		}
	default:
		return tools.Result{Content: fmt.Sprintf("Invalid edit_mode: %s", editMode), IsError: true}, nil
	}

	// Write the modified notebook back
	out, err := json.MarshalIndent(nb, "", "  ")
	if err != nil {
		return tools.Result{Content: fmt.Sprintf("Cannot serialize notebook: %v", err), IsError: true}, nil
	}
	out = append(out, '\n')

	if err := os.WriteFile(path, out, 0644); err != nil {
		return tools.Result{Content: fmt.Sprintf("Cannot write notebook: %v", err), IsError: true}, nil
	}

	return tools.Result{Content: fmt.Sprintf("Notebook edited: %s", path)}, nil
}

// newNotebook creates a minimal empty notebook structure.
func newNotebook() notebook {
	return notebook{
		NbFormat:      4,
		NbFormatMinor: 5,
		Metadata:      map[string]any{},
		Cells:         []notebookCell{},
	}
}

// replaceCell replaces the source of the cell at the given cell_number.
func replaceCell(nb *notebook, input notebookEditInput) error {
	if input.CellNumber == nil {
		return fmt.Errorf("cell_number is required for replace mode")
	}
	idx := *input.CellNumber
	if idx < 0 || idx >= len(nb.Cells) {
		return fmt.Errorf("cell_number %d is out of range (notebook has %d cells)", idx, len(nb.Cells))
	}
	nb.Cells[idx].Source = input.NewSource
	return nil
}

// insertCell inserts a new cell at the position specified by cell_number,
// or appends at the end if cell_number is not provided.
func insertCell(nb *notebook, input notebookEditInput) {
	cellType := input.CellType
	if cellType == "" {
		cellType = "code"
	}

	cell := notebookCell{
		CellType: cellType,
		Metadata: map[string]any{},
		Source:   input.NewSource,
	}

	if input.CellNumber == nil {
		nb.Cells = append(nb.Cells, cell)
		return
	}

	idx := max(0, min(*input.CellNumber, len(nb.Cells)))

	nb.Cells = append(nb.Cells[:idx], append([]notebookCell{cell}, nb.Cells[idx:]...)...)
}

// deleteCell removes the cell at the given cell_number.
func deleteCell(nb *notebook, input notebookEditInput) error {
	if input.CellNumber == nil {
		return fmt.Errorf("cell_number is required for delete mode")
	}
	idx := *input.CellNumber
	if idx < 0 || idx >= len(nb.Cells) {
		return fmt.Errorf("cell_number %d is out of range (notebook has %d cells)", idx, len(nb.Cells))
	}

	nb.Cells = append(nb.Cells[:idx], nb.Cells[idx+1:]...)
	return nil
}
