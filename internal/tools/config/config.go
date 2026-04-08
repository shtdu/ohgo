// Package config implements the config tool for viewing settings.
package config

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/shtdu/ohgo/internal/config"
	"github.com/shtdu/ohgo/internal/tools"
)

type configInput struct {
	Action string `json:"action"`
	Key    string `json:"key,omitempty"`
}

// ConfigTool shows the current configuration settings.
type ConfigTool struct {
	Settings *config.Settings
}

func (ConfigTool) Name() string { return "config" }

func (ConfigTool) Description() string {
	return "View current configuration settings. Use 'show' action to display all settings or a specific key."
}

func (ConfigTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"description": "The action to perform",
				"enum":        []string{"show"},
			},
			"key": map[string]any{
				"type":        "string",
				"description": "Optional specific setting key to show (e.g. 'model', 'max_tokens')",
			},
		},
		"required":             []string{"action"},
		"additionalProperties": false,
	}
}

func (t ConfigTool) Execute(_ context.Context, args json.RawMessage) (tools.Result, error) {
	var input configInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{
			Content: fmt.Sprintf("invalid arguments: %v", err),
			IsError: true,
		}, nil
	}

	if input.Action != "show" {
		return tools.Result{
			Content: fmt.Sprintf("unknown action: %q (supported: show)", input.Action),
			IsError: true,
		}, nil
	}

	if t.Settings == nil {
		return tools.Result{
			Content: "config settings are not available",
			IsError: true,
		}, nil
	}

	if input.Key == "" {
		data, err := json.MarshalIndent(t.Settings, "", "  ")
		if err != nil {
			return tools.Result{
				Content: fmt.Sprintf("failed to marshal settings: %v", err),
				IsError: true,
			}, nil
		}
		return tools.Result{Content: string(data)}, nil
	}

	// Look up a specific field by JSON tag name.
	val, err := lookupField(t.Settings, input.Key)
	if err != nil {
		return tools.Result{
			Content: err.Error(),
			IsError: true,
		}, nil
	}

	data, err := json.MarshalIndent(val, "", "  ")
	if err != nil {
		return tools.Result{
			Content: fmt.Sprintf("failed to marshal value: %v", err),
			IsError: true,
		}, nil
	}
	return tools.Result{Content: string(data)}, nil
}

// lookupField finds a field in the Settings struct by its JSON tag name.
func lookupField(s *config.Settings, key string) (any, error) {
	v := reflect.ValueOf(s).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")
		// Strip omitempty and other options from the tag.
		name := tag
		if idx := indexByte(tag, ','); idx >= 0 {
			name = tag[:idx]
		}
		if name == key {
			return v.Field(i).Interface(), nil
		}
	}
	return nil, fmt.Errorf("unknown setting key: %q", key)
}

// indexByte returns the index of the first occurrence of sep in s,
// or -1 if not present.
func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
