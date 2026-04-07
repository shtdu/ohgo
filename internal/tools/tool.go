// Package tools defines the Tool interface and registry for agent tool execution.
package tools

import (
	"context"
	"encoding/json"
)

// Result represents the output of a tool execution.
type Result struct {
	Content string
	IsError bool
}

// Tool is the interface all agent tools must implement.
// It mirrors the Python BaseTool pattern with JSON Schema support.
type Tool interface {
	// Name returns the unique tool identifier (e.g. "bash", "read_file").
	Name() string

	// Description returns a human-readable description of what the tool does.
	Description() string

	// InputSchema returns the JSON Schema for the tool's input parameters.
	InputSchema() map[string]any

	// Execute runs the tool with the given JSON-encoded arguments.
	Execute(ctx context.Context, args json.RawMessage) (Result, error)
}

// Registry manages available tools.
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

// Register adds a tool to the registry.
func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

// Get retrieves a tool by name. Returns nil if not found.
func (r *Registry) Get(name string) Tool {
	return r.tools[name]
}

// List returns all registered tools.
func (r *Registry) List() []Tool {
	out := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t)
	}
	return out
}
