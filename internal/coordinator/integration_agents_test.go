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

// EARS: REQ-AC-002
func TestIntegration_Coordinator_CreateTeam(t *testing.T) {
	c := coordinator.New("/bin/echo")

	err := c.CreateTeam("alpha", []string{"agent1", "agent2"})
	require.NoError(t, err)

	teams := c.ListTeams()
	require.Len(t, teams, 1)
	assert.Equal(t, "alpha", teams[0].Name)
	assert.Contains(t, teams[0].Agents, "agent1")
	assert.Contains(t, teams[0].Agents, "agent2")
}

// EARS: REQ-AC-002
func TestIntegration_Coordinator_DuplicateTeamError(t *testing.T) {
	c := coordinator.New("/bin/echo")

	err := c.CreateTeam("beta", nil)
	require.NoError(t, err)

	err = c.CreateTeam("beta", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

// EARS: REQ-AC-002
func TestIntegration_Coordinator_DeleteTeam(t *testing.T) {
	c := coordinator.New("/bin/echo")

	err := c.CreateTeam("gamma", nil)
	require.NoError(t, err)

	err = c.DeleteTeam("gamma")
	require.NoError(t, err)

	teams := c.ListTeams()
	assert.Empty(t, teams)
}

// EARS: REQ-AC-002
func TestIntegration_Coordinator_DeleteNonExistentTeam(t *testing.T) {
	c := coordinator.New("/bin/echo")

	err := c.DeleteTeam("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// EARS: REQ-AC-001
func TestIntegration_Coordinator_SpawnEcho(t *testing.T) {
	c := coordinator.New("/bin/echo")
	defer c.Shutdown()

	agentID, err := c.Spawn(context.Background(), coordinator.AgentSpec{
		Name:        "echo-agent",
		Description: "echo test agent",
		Prompt:      "hello from agent",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, agentID)

	// Verify agent is tracked
	ra, found := c.Get(agentID)
	require.True(t, found)
	assert.Equal(t, "echo-agent", ra.Name)
	assert.Equal(t, coordinator.AgentStatusRunning, ra.Status)
	assert.Greater(t, ra.PID, 0)

	// Wait for the echo command to finish
	time.Sleep(500 * time.Millisecond)

	ra, found = c.Get(agentID)
	require.True(t, found)
	assert.Equal(t, coordinator.AgentStatusDone, ra.Status)
}

// EARS: REQ-AC-004
func TestIntegration_Coordinator_AgentIsolation(t *testing.T) {
	c := coordinator.New("/bin/echo")
	defer c.Shutdown()

	id1, err := c.Spawn(context.Background(), coordinator.AgentSpec{
		Name: "agent-1", Prompt: "prompt 1",
	})
	require.NoError(t, err)

	id2, err := c.Spawn(context.Background(), coordinator.AgentSpec{
		Name: "agent-2", Prompt: "prompt 2",
	})
	require.NoError(t, err)

	// Agents have separate IDs
	assert.NotEqual(t, id1, id2)

	ra1, _ := c.Get(id1)
	ra2, _ := c.Get(id2)
	assert.Equal(t, "agent-1", ra1.Name)
	assert.Equal(t, "agent-2", ra2.Name)
	assert.NotEqual(t, ra1.PID, ra2.PID, "agents should have separate processes")
}

// EARS: REQ-AC-001
func TestIntegration_Coordinator_ListAgents(t *testing.T) {
	c := coordinator.New("/bin/echo")
	defer c.Shutdown()

	_, err := c.Spawn(context.Background(), coordinator.AgentSpec{Name: "a1", Prompt: "p1"})
	require.NoError(t, err)
	_, err = c.Spawn(context.Background(), coordinator.AgentSpec{Name: "a2", Prompt: "p2"})
	require.NoError(t, err)

	agents := c.List()
	assert.GreaterOrEqual(t, len(agents), 2)
}
