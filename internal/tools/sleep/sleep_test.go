package sleep

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/shtdu/ohgo/internal/tools"
)

func TestSleepTool_Name(t *testing.T) {
	assert.Equal(t, "sleep", SleepTool{}.Name())
}

func TestSleepTool_ShortSleep(t *testing.T) {
	tool := SleepTool{}
	args, _ := json.Marshal(map[string]float64{"seconds": 0.05})

	start := time.Now()
	result, err := tool.Execute(context.Background(), args)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Slept for")
	assert.GreaterOrEqual(t, elapsed.Seconds(), 0.04)
}

func TestSleepTool_ContextCancel(t *testing.T) {
	tool := SleepTool{}
	args, _ := json.Marshal(map[string]float64{"seconds": 30.0})

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err := tool.Execute(ctx, args)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Less(t, elapsed.Seconds(), 2.0)
}

func TestSleepTool_ClampsToUpperBound(t *testing.T) {
	tool := SleepTool{}
	args, _ := json.Marshal(map[string]float64{"seconds": 100.0})

	start := time.Now()
	result, err := tool.Execute(context.Background(), args)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Slept for 30.0 seconds")
	assert.Less(t, elapsed.Seconds(), 31.0)
}

func TestSleepTool_ClampsNegativeToZero(t *testing.T) {
	tool := SleepTool{}
	args, _ := json.Marshal(map[string]float64{"seconds": -5.0})

	start := time.Now()
	result, err := tool.Execute(context.Background(), args)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Slept for 0.0 seconds")
	assert.Less(t, elapsed.Seconds(), 1.0)
}

func TestSleepTool_InvalidJSON(t *testing.T) {
	tool := SleepTool{}
	result, err := tool.Execute(context.Background(), json.RawMessage(`{bad`))
	require.NoError(t, err)
	assert.True(t, result.IsError)
}

func TestSleepTool_DefaultSeconds(t *testing.T) {
	tool := SleepTool{}
	// Pass empty object — should use default of 1.0
	args := json.RawMessage(`{}`)

	start := time.Now()
	result, err := tool.Execute(context.Background(), args)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.False(t, result.IsError)
	assert.Contains(t, result.Content, "Slept for 1.0 seconds")
	assert.GreaterOrEqual(t, elapsed.Seconds(), 0.9)
}

var _ tools.Tool = SleepTool{}
