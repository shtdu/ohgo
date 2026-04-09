package permissions

// Mode represents a permission mode (e.g. "default", "plan", "auto").
type Mode string

const (
	ModeDefault Mode = "default"
	ModePlan    Mode = "plan"
	ModeAuto    Mode = "auto"
)

// ToolCategory classifies a tool as read-only or mutating.
type ToolCategory int

const (
	// CategoryRead indicates a safe, non-destructive tool.
	CategoryRead ToolCategory = iota
	// CategoryWrite indicates a tool that modifies state.
	CategoryWrite
)

// ParseMode converts a string to a Mode, returning ModeDefault for unknown values.
func ParseMode(s string) Mode {
	switch Mode(s) {
	case ModeDefault:
		return ModeDefault
	case ModePlan:
		return ModePlan
	case ModeAuto:
		return ModeAuto
	default:
		return ModeDefault
	}
}

// ClassifyTool returns the category for a given tool name.
// Known read tools: read_file, glob, grep, web_fetch, web_search, lsp.
// All other tools default to CategoryWrite (safe default).
func ClassifyTool(toolName string) ToolCategory {
	readTools := map[string]bool{
		"read_file":          true,
		"glob":               true,
		"grep":               true,
		"web_fetch":          true,
		"web_search":         true,
		"lsp":                true,
		"mcp_list_resources": true,
		"mcp_read_resource":  true,
		"mcp_auth":           true,
	}
	if readTools[toolName] {
		return CategoryRead
	}
	return CategoryWrite
}
