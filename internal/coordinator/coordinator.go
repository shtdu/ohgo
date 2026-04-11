// Package coordinator handles multi-agent subagent spawning and team coordination.
package coordinator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

// AgentSpec describes a subagent to spawn.
type AgentSpec struct {
	Name        string
	Description string
	Prompt      string
}

// AgentStatus represents the current state of a running agent.
type AgentStatus string

const (
	AgentStatusStarting AgentStatus = "starting"
	AgentStatusRunning  AgentStatus = "running"
	AgentStatusDone     AgentStatus = "done"
	AgentStatusFailed   AgentStatus = "failed"
)

// RunningAgent tracks a live subagent process.
type RunningAgent struct {
	ID        string
	Name      string
	Status    AgentStatus
	PID       int
	StartedAt time.Time
}

// Team groups named agent definitions for coordinated work.
type Team struct {
	Name      string    `json:"name"`
	Agents    []string  `json:"agents"`
	CreatedAt time.Time `json:"created_at"`
}

// Coordinator manages multi-agent orchestration, including spawning,
// tracking, and stopping subagent processes as well as team management.
type Coordinator struct {
	mu      sync.RWMutex
	agents  map[string]*RunningAgent
	procs   map[string]*os.Process
	cancels map[string]context.CancelFunc
	teams   map[string]*Team
	defs    map[string]*AgentDef
	binPath string
}

// New creates a new Coordinator that uses binPath as the executable for
// spawned subagents.
func New(binPath string) *Coordinator {
	return &Coordinator{
		agents:  make(map[string]*RunningAgent),
		procs:   make(map[string]*os.Process),
		cancels: make(map[string]context.CancelFunc),
		teams:   make(map[string]*Team),
		defs:    make(map[string]*AgentDef),
		binPath: binPath,
	}
}

// RegisterDefs stores agent definitions by name for later reference.
func (c *Coordinator) RegisterDefs(defs []*AgentDef) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, d := range defs {
		c.defs[d.Name] = d
	}
}

// Spawn launches a subagent process using the configured binPath.
// It returns the unique agent ID or an error if the process cannot be started.
func (c *Coordinator) Spawn(ctx context.Context, spec AgentSpec) (string, error) {
	agentID := fmt.Sprintf("agent-%d", time.Now().UnixNano())

	ctx, cancel := context.WithCancel(ctx)

	cmd := exec.CommandContext(ctx, c.binPath, "--prompt", spec.Prompt)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		cancel()
		return "", fmt.Errorf("start agent %q: %w", spec.Name, err)
	}

	ra := &RunningAgent{
		ID:        agentID,
		Name:      spec.Name,
		Status:    AgentStatusRunning,
		PID:       cmd.Process.Pid,
		StartedAt: time.Now(),
	}

	c.mu.Lock()
	c.agents[agentID] = ra
	c.procs[agentID] = cmd.Process
	c.cancels[agentID] = cancel
	c.mu.Unlock()

	// Monitor the process in the background and update status.
	// Only update if still running — Stop/Shutdown may have already set it.
	go func() {
		err := cmd.Wait()
		c.mu.Lock()
		defer c.mu.Unlock()
		if ra.Status != AgentStatusRunning {
			// Already stopped via Stop/Shutdown; clean up map entries.
			delete(c.procs, agentID)
			delete(c.cancels, agentID)
			return
		}
		if err != nil {
			ra.Status = AgentStatusFailed
		} else {
			ra.Status = AgentStatusDone
		}
		// Clean up completed agent entries.
		delete(c.procs, agentID)
		delete(c.cancels, agentID)
	}()

	return agentID, nil
}

// Stop terminates a running agent by its ID.
func (c *Coordinator) Stop(_ context.Context, agentID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	proc, ok := c.procs[agentID]
	if !ok {
		return fmt.Errorf("agent %q not found", agentID)
	}

	if cancel, ok := c.cancels[agentID]; ok {
		cancel()
	}

	if err := proc.Kill(); err != nil && err != os.ErrProcessDone {
		return fmt.Errorf("kill agent %q: %w", agentID, err)
	}

	if ra, ok := c.agents[agentID]; ok {
		ra.Status = AgentStatusDone
	}

	return nil
}

// Get returns a snapshot of a running agent by ID.
// The returned value is a copy and safe to inspect without holding the lock.
func (c *Coordinator) Get(agentID string) (RunningAgent, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ra, ok := c.agents[agentID]
	if !ok {
		return RunningAgent{}, false
	}
	return *ra, true
}

// List returns snapshots of all tracked agents.
// The returned values are copies and safe to inspect without holding the lock.
func (c *Coordinator) List() []RunningAgent {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make([]RunningAgent, 0, len(c.agents))
	for _, ra := range c.agents {
		out = append(out, *ra)
	}
	return out
}

// Shutdown stops all running agents and cleans up resources.
func (c *Coordinator) Shutdown() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for id, proc := range c.procs {
		if cancel, ok := c.cancels[id]; ok {
			cancel()
		}
		_ = proc.Kill()
		if ra, ok := c.agents[id]; ok {
			ra.Status = AgentStatusDone
		}
	}

	// Clear all maps to release references.
	c.procs = make(map[string]*os.Process)
	c.cancels = make(map[string]context.CancelFunc)
}

// CreateTeam creates a named team of agent definitions.
// Returns an error if a team with the same name already exists.
func (c *Coordinator) CreateTeam(name string, agentNames []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.teams[name]; exists {
		return fmt.Errorf("team %q already exists", name)
	}

	c.teams[name] = &Team{
		Name:      name,
		Agents:    agentNames,
		CreatedAt: time.Now(),
	}
	return nil
}

// DeleteTeam removes a team by name.
// Returns an error if the team does not exist.
func (c *Coordinator) DeleteTeam(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.teams[name]; !exists {
		return fmt.Errorf("team %q not found", name)
	}

	delete(c.teams, name)
	return nil
}

// ListTeams returns all defined teams.
func (c *Coordinator) ListTeams() []*Team {
	c.mu.RLock()
	defer c.mu.RUnlock()

	out := make([]*Team, 0, len(c.teams))
	for _, t := range c.teams {
		out = append(out, t)
	}
	return out
}
