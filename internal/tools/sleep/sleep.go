// Package sleep implements the sleep tool for pausing execution.
package sleep

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shtdu/ohgo/internal/tools"
)

const (
	minSleepSeconds = 0.0
	maxSleepSeconds = 30.0
	defaultSeconds  = 1.0
)

type sleepInput struct {
	Seconds float64 `json:"seconds"`
}

// SleepTool pauses execution for a specified duration.
type SleepTool struct{}

func (SleepTool) Name() string { return "sleep" }

func (SleepTool) Description() string {
	return "Pause execution for a specified number of seconds. Useful for waiting between operations."
}

func (SleepTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"seconds": map[string]any{
				"type":        "number",
				"description": "Number of seconds to sleep",
				"default":     defaultSeconds,
				"minimum":     minSleepSeconds,
				"maximum":     maxSleepSeconds,
			},
		},
		"additionalProperties": false,
	}
}

func (SleepTool) Execute(ctx context.Context, args json.RawMessage) (tools.Result, error) {
	var input sleepInput
	if err := json.Unmarshal(args, &input); err != nil {
		return tools.Result{Content: fmt.Sprintf("invalid arguments: %v", err), IsError: true}, nil
	}

	// Apply default
	seconds := input.Seconds
	if seconds == 0 && len(args) > 0 {
		// Check if seconds was explicitly provided as 0
		var raw map[string]json.RawMessage
		if json.Unmarshal(args, &raw) == nil {
			if _, ok := raw["seconds"]; !ok {
				seconds = defaultSeconds
			}
		}
	} else if seconds == 0 {
		seconds = defaultSeconds
	}

	// Clamp to valid range
	if seconds < minSleepSeconds {
		seconds = minSleepSeconds
	}
	if seconds > maxSleepSeconds {
		seconds = maxSleepSeconds
	}

	d := time.Duration(seconds * float64(time.Second))

	timer := time.NewTimer(d)
	select {
	case <-timer.C:
		return tools.Result{Content: fmt.Sprintf("Slept for %.1f seconds", d.Seconds())}, nil
	case <-ctx.Done():
		timer.Stop()
		return tools.Result{}, ctx.Err()
	}
}
