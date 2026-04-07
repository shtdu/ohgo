package engine

import (
	"sync"

	"github.com/shtdu/ohgo/internal/api"
)

// CostTracker accumulates token usage across a session.
type CostTracker struct {
	mu       sync.RWMutex
	usage    api.UsageSnapshot
	turns    int
}

// NewCostTracker creates a new cost tracker.
func NewCostTracker() *CostTracker {
	return &CostTracker{}
}

// Add accumulates a usage snapshot.
func (t *CostTracker) Add(u api.UsageSnapshot) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.usage.InputTokens += u.InputTokens
	t.usage.OutputTokens += u.OutputTokens
}

// Total returns the aggregated usage.
func (t *CostTracker) Total() api.UsageSnapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.usage
}

// Turns returns the number of turns tracked.
func (t *CostTracker) Turns() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.turns
}

// IncrementTurns increments the turn counter.
func (t *CostTracker) IncrementTurns() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.turns++
}

// Reset clears all tracking.
func (t *CostTracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.usage = api.UsageSnapshot{}
	t.turns = 0
}
