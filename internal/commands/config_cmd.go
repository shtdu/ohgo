package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// configCmd shows the current configuration as JSON.
type configCmd struct{}

var _ Command = configCmd{}

func (configCmd) Name() string     { return "config" }
func (configCmd) ShortHelp() string { return "show current configuration" }

func (configCmd) Run(_ context.Context, _ string, deps *Deps) (Result, error) {
	if deps.Config == nil {
		return Result{Output: "config: no configuration loaded"}, nil
	}

	// Marshal a copy with the API key masked.
	data, err := json.MarshalIndent(deps.Config, "", "  ")
	if err != nil {
		return Result{}, fmt.Errorf("config: marshal: %w", err)
	}

	// Mask sensitive fields in the output.
	s := string(data)
	s = maskJSONValue(s, `"api_key"`, deps.Config.APIKey)

	return Result{Output: s}, nil
}

// maskJSONValue replaces a sensitive value in JSON output with "****".
func maskJSONValue(jsonStr, _ string, value string) string {
	if value == "" {
		return jsonStr
	}
	return strings.Replace(jsonStr, `"`+value+`"`, `"****"`, 1)
}
