package permissions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseMode(t *testing.T) {
	tests := []struct {
		input string
		want  Mode
	}{
		{"default", ModeDefault},
		{"plan", ModePlan},
		{"auto", ModeAuto},
		{"", ModeDefault},
		{"unknown", ModeDefault},
		{"DEFAULT", ModeDefault},
		{"Auto", ModeDefault},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, ParseMode(tt.input))
		})
	}
}

func TestClassifyTool(t *testing.T) {
	tests := []struct {
		name string
		want ToolCategory
	}{
		{"read_file", CategoryRead},
		{"glob", CategoryRead},
		{"grep", CategoryRead},
		{"write_file", CategoryWrite},
		{"edit_file", CategoryWrite},
		{"bash", CategoryWrite},
		{"agent", CategoryWrite},
		{"unknown_tool", CategoryWrite},
		{"", CategoryWrite},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ClassifyTool(tt.name))
		})
	}
}
