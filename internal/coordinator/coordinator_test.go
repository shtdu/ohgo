package coordinator

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	c := New("/usr/bin/echo")
	require.NotNil(t, c)
	assert.Equal(t, "/usr/bin/echo", c.binPath)
}

func TestRegisterDefs(t *testing.T) {
	c := New("/bin/true")

	defs := []*AgentDef{
		{Name: "agent-a", Description: "Agent A", Prompt: "Do A"},
		{Name: "agent-b", Description: "Agent B", Prompt: "Do B"},
	}
	c.RegisterDefs(defs)

	c.mu.RLock()
	defer c.mu.RUnlock()
	assert.Contains(t, c.defs, "agent-a")
	assert.Contains(t, c.defs, "agent-b")
	assert.Equal(t, "Do A", c.defs["agent-a"].Prompt)
}

func TestCreateTeam(t *testing.T) {
	c := New("/bin/true")

	err := c.CreateTeam("research", []string{"agent-a", "agent-b"})
	require.NoError(t, err)

	teams := c.ListTeams()
	require.Len(t, teams, 1)
	assert.Equal(t, "research", teams[0].Name)
	assert.Equal(t, []string{"agent-a", "agent-b"}, teams[0].Agents)
}

func TestCreateTeam_DuplicateError(t *testing.T) {
	c := New("/bin/true")

	err := c.CreateTeam("research", []string{"agent-a"})
	require.NoError(t, err)

	err = c.CreateTeam("research", []string{"agent-b"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestDeleteTeam(t *testing.T) {
	c := New("/bin/true")

	err := c.CreateTeam("research", []string{"agent-a"})
	require.NoError(t, err)

	err = c.DeleteTeam("research")
	require.NoError(t, err)

	teams := c.ListTeams()
	assert.Empty(t, teams)
}

func TestDeleteTeam_NotFoundError(t *testing.T) {
	c := New("/bin/true")

	err := c.DeleteTeam("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListTeams_Multiple(t *testing.T) {
	c := New("/bin/true")

	require.NoError(t, c.CreateTeam("alpha", []string{"a1"}))
	require.NoError(t, c.CreateTeam("beta", []string{"b1"}))

	teams := c.ListTeams()
	assert.Len(t, teams, 2)

	names := map[string]bool{}
	for _, tm := range teams {
		names[tm.Name] = true
	}
	assert.True(t, names["alpha"])
	assert.True(t, names["beta"])
}

func TestList_Empty(t *testing.T) {
	c := New("/bin/true")
	agents := c.List()
	assert.Empty(t, agents)
}

func TestSpawn_EchoCompletes(t *testing.T) {
	c := New("/bin/echo")

	agentID, err := c.Spawn(context.Background(), AgentSpec{
		Name:   "test-echo",
		Prompt: "hello",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, agentID)

	ra, ok := c.Get(agentID)
	require.True(t, ok)
	assert.Equal(t, "test-echo", ra.Name)
	assert.Greater(t, ra.PID, 0)

	// Wait for the process to complete since echo exits immediately.
	assert.Eventually(t, func() bool {
		snapshot, _ := c.Get(agentID)
		return snapshot.Status == AgentStatusDone
	}, 3*time.Second, 50*time.Millisecond, "agent should reach done status")
}

func TestSpawn_Stop(t *testing.T) {
	// Use sleep so the process stays alive long enough to stop it.
	c := New("/bin/sleep")

	agentID, err := c.Spawn(context.Background(), AgentSpec{
		Name:   "test-sleep",
		Prompt: "300",
	})
	require.NoError(t, err)

	ra, ok := c.Get(agentID)
	require.True(t, ok)
	assert.Equal(t, AgentStatusRunning, ra.Status)

	err = c.Stop(context.Background(), agentID)
	require.NoError(t, err)

	// Wait briefly for the goroutine to observe the process exit.
	assert.Eventually(t, func() bool {
		snapshot, _ := c.Get(agentID)
		return snapshot.Status == AgentStatusDone
	}, 2*time.Second, 50*time.Millisecond, "agent should reach done status after stop")
}

func TestShutdown(t *testing.T) {
	c := New("/bin/sleep")

	_, err := c.Spawn(context.Background(), AgentSpec{
		Name:   "agent-1",
		Prompt: "300",
	})
	require.NoError(t, err)

	_, err = c.Spawn(context.Background(), AgentSpec{
		Name:   "agent-2",
		Prompt: "300",
	})
	require.NoError(t, err)

	assert.Len(t, c.List(), 2)

	c.Shutdown()

	// Wait briefly for goroutines to observe the killed processes.
	assert.Eventually(t, func() bool {
		for _, ra := range c.List() {
			if ra.Status != AgentStatusDone {
				return false
			}
		}
		return true
	}, 2*time.Second, 50*time.Millisecond, "all agents should reach done status after shutdown")
}

func TestGet_NotFound(t *testing.T) {
	c := New("/bin/true")
	_, ok := c.Get("nonexistent")
	assert.False(t, ok)
}

func TestStop_NotFound(t *testing.T) {
	c := New("/bin/true")
	err := c.Stop(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
