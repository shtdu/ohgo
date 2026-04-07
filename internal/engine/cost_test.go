package engine

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/shtdu/ohgo/internal/api"
)

func TestCostTracker_Add(t *testing.T) {
	tracker := NewCostTracker()
	tracker.Add(api.UsageSnapshot{InputTokens: 100, OutputTokens: 50})
	tracker.Add(api.UsageSnapshot{InputTokens: 200, OutputTokens: 75})

	total := tracker.Total()
	assert.Equal(t, 300, total.InputTokens)
	assert.Equal(t, 125, total.OutputTokens)
}

func TestCostTracker_Turns(t *testing.T) {
	tracker := NewCostTracker()
	assert.Equal(t, 0, tracker.Turns())

	tracker.IncrementTurns()
	tracker.IncrementTurns()
	assert.Equal(t, 2, tracker.Turns())
}

func TestCostTracker_Reset(t *testing.T) {
	tracker := NewCostTracker()
	tracker.Add(api.UsageSnapshot{InputTokens: 100, OutputTokens: 50})
	tracker.IncrementTurns()

	tracker.Reset()
	assert.Equal(t, 0, tracker.Total().InputTokens)
	assert.Equal(t, 0, tracker.Turns())
}

func TestCostTracker_Concurrent(t *testing.T) {
	tracker := NewCostTracker()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tracker.Add(api.UsageSnapshot{InputTokens: 1, OutputTokens: 1})
		}()
	}
	wg.Wait()

	total := tracker.Total()
	assert.Equal(t, 100, total.InputTokens)
	assert.Equal(t, 100, total.OutputTokens)
}
