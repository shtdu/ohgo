package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsageSnapshot_TotalTokens(t *testing.T) {
	u := UsageSnapshot{InputTokens: 100, OutputTokens: 50}
	assert.Equal(t, 150, u.TotalTokens())
}

func TestUsageSnapshot_ZeroValue(t *testing.T) {
	var u UsageSnapshot
	assert.Equal(t, 0, u.TotalTokens())
}

func TestUsageSnapshot_JSONRoundTrip(t *testing.T) {
	original := UsageSnapshot{InputTokens: 200, OutputTokens: 100}
	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded UsageSnapshot
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, 200, decoded.InputTokens)
	assert.Equal(t, 100, decoded.OutputTokens)
	assert.Equal(t, 300, decoded.TotalTokens())
}
