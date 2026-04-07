// Package coordinator handles multi-agent subagent spawning and team coordination.
package coordinator

import (
	"context"
)

// AgentSpec describes a subagent to spawn.
type AgentSpec struct {
	Name        string
	Description string
	Prompt      string
}

// Coordinator manages multi-agent orchestration.
type Coordinator struct{}

// New creates a new coordinator.
func New() *Coordinator {
	return &Coordinator{}
}

// Spawn launches a subagent with the given spec.
func (c *Coordinator) Spawn(ctx context.Context, spec AgentSpec) error {
	// TODO: implement subagent spawning
	return nil
}
