//go:build integration

package coordinator_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shtdu/ohgo/internal/coordinator"
)

// EARS: REQ-AC-001
// Spawn an agent and verify its full lifecycle:
// starts as running → transitions to done after process exits.
func TestIntegration_Coordinator_SpawnLifecycle(t *testing.T) {
	c := coordinator.New("/bin/echo")
	defer c.Shutdown()

	agentID, err := c.Spawn(context.Background(), coordinator.AgentSpec{
		Name:        "lifecycle-agent",
		Description: "lifecycle test",
		Prompt:      "hello",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, agentID)

	// Immediately: should be tracked as running
	ra, found := c.Get(agentID)
	require.True(t, found)
	assert.Equal(t, "lifecycle-agent", ra.Name)
	assert.Equal(t, coordinator.AgentStatusRunning, ra.Status)
	assert.Greater(t, ra.PID, 0)
	assert.WithinDuration(t, time.Now(), ra.StartedAt, 2*time.Second)

	// Wait for echo to complete
	time.Sleep(500 * time.Millisecond)

	ra, found = c.Get(agentID)
	require.True(t, found)
	assert.Equal(t, coordinator.AgentStatusDone, ra.Status)
}

// EARS: REQ-AC-004
// Multiple spawned agents have separate PIDs, names, and IDs.
func TestIntegration_Coordinator_AgentIsolation(t *testing.T) {
	c := coordinator.New("/bin/echo")
	defer c.Shutdown()

	id1, err := c.Spawn(context.Background(), coordinator.AgentSpec{Name: "agent-1", Prompt: "p1"})
	require.NoError(t, err)
	id2, err := c.Spawn(context.Background(), coordinator.AgentSpec{Name: "agent-2", Prompt: "p2"})
	require.NoError(t, err)

	assert.NotEqual(t, id1, id2, "agents must have unique IDs")

	ra1, _ := c.Get(id1)
	ra2, _ := c.Get(id2)
	assert.Equal(t, "agent-1", ra1.Name)
	assert.Equal(t, "agent-2", ra2.Name)
	assert.NotEqual(t, ra1.PID, ra2.PID, "agents must run in separate processes")
}

// EARS: REQ-AC-001
// List returns all spawned agents with correct count and names.
func TestIntegration_Coordinator_ListAgents(t *testing.T) {
	c := coordinator.New("/bin/echo")
	defer c.Shutdown()

	_, err := c.Spawn(context.Background(), coordinator.AgentSpec{Name: "a1", Prompt: "p1"})
	require.NoError(t, err)
	_, err = c.Spawn(context.Background(), coordinator.AgentSpec{Name: "a2", Prompt: "p2"})
	require.NoError(t, err)

	agents := c.List()
	assert.GreaterOrEqual(t, len(agents), 2)

	names := make(map[string]bool)
	for _, a := range agents {
		names[a.Name] = true
	}
	assert.True(t, names["a1"])
	assert.True(t, names["a2"])
}

// EARS: REQ-AC-001
// Stop kills a running agent, transitioning it to done status.
func TestIntegration_Coordinator_StopRunningAgent(t *testing.T) {
	c := coordinator.New("/bin/sleep")
	defer c.Shutdown()

	agentID, err := c.Spawn(context.Background(), coordinator.AgentSpec{
		Name:   "long-runner",
		Prompt: "300", // sleep 300 seconds
	})
	require.NoError(t, err)

	// Verify running
	ra, found := c.Get(agentID)
	require.True(t, found)
	assert.Equal(t, coordinator.AgentStatusRunning, ra.Status)

	// Stop it
	require.NoError(t, c.Stop(context.Background(), agentID))

	// Should be done now
	ra, found = c.Get(agentID)
	require.True(t, found)
	assert.Equal(t, coordinator.AgentStatusDone, ra.Status)
}

// EARS: REQ-AC-002
// Team CRUD: create team, list teams, create with agents, verify membership.
func TestIntegration_Coordinator_TeamWithAgents(t *testing.T) {
	c := coordinator.New("/bin/echo")

	// Create team with named agents
	require.NoError(t, c.CreateTeam("deploy-team", []string{"builder", "tester"}))

	teams := c.ListTeams()
	require.Len(t, teams, 1)
	assert.Equal(t, "deploy-team", teams[0].Name)
	assert.Contains(t, teams[0].Agents, "builder")
	assert.Contains(t, teams[0].Agents, "tester")

	// Duplicate should error
	err := c.CreateTeam("deploy-team", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Delete and verify empty
	require.NoError(t, c.DeleteTeam("deploy-team"))
	assert.Empty(t, c.ListTeams())

	// Delete nonexistent should error
	err = c.DeleteTeam("deploy-team")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// EARS: REQ-AC-002, REQ-AC-001
// Spawning agents that belong to a team and tracking them together.
func TestIntegration_Coordinator_TeamAgentTracking(t *testing.T) {
	c := coordinator.New("/bin/echo")
	defer c.Shutdown()

	// Define a team
	require.NoError(t, c.CreateTeam("workers", []string{"worker-a", "worker-b"}))

	// Spawn agents with matching names
	idA, err := c.Spawn(context.Background(), coordinator.AgentSpec{Name: "worker-a", Prompt: "work a"})
	require.NoError(t, err)
	idB, err := c.Spawn(context.Background(), coordinator.AgentSpec{Name: "worker-b", Prompt: "work b"})
	require.NoError(t, err)

	// Both tracked
	agents := c.List()
	assert.GreaterOrEqual(t, len(agents), 2)

	// Team definition still intact
	teams := c.ListTeams()
	require.Len(t, teams, 1)
	assert.Equal(t, "workers", teams[0].Name)
	assert.Len(t, teams[0].Agents, 2)

	// Wait for agents to finish
	time.Sleep(500 * time.Millisecond)

	// Verify both completed
	raA, _ := c.Get(idA)
	raB, _ := c.Get(idB)
	assert.Equal(t, coordinator.AgentStatusDone, raA.Status)
	assert.Equal(t, coordinator.AgentStatusDone, raB.Status)
}
