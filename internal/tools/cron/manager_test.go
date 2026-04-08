package cron

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	mgr := NewManager()
	assert.NotNil(t, mgr)
	assert.Empty(t, mgr.List())
}

func TestManager_CreateAndList(t *testing.T) {
	mgr := NewManager()
	err := mgr.Create(Job{
		Name:     "test-job",
		Schedule: "0 * * * *",
		Command:  "echo hello",
		Enabled:  true,
	})
	require.NoError(t, err)

	jobs := mgr.List()
	assert.Len(t, jobs, 1)
	assert.Equal(t, "test-job", jobs[0].Name)
	assert.Equal(t, "0 * * * *", jobs[0].Schedule)
	assert.True(t, jobs[0].Enabled)
	assert.False(t, jobs[0].CreatedAt.IsZero())
}

func TestManager_CreateInvalidSchedule(t *testing.T) {
	mgr := NewManager()
	err := mgr.Create(Job{
		Name:     "bad-job",
		Schedule: "not-a-cron",
		Command:  "echo fail",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid schedule")
	assert.Empty(t, mgr.List())
}

func TestManager_CreateUpsert(t *testing.T) {
	mgr := NewManager()
	err := mgr.Create(Job{Name: "job1", Schedule: "0 * * * *", Command: "v1"})
	require.NoError(t, err)
	err = mgr.Create(Job{Name: "job1", Schedule: "0 0 * * *", Command: "v2"})
	require.NoError(t, err)
	jobs := mgr.List()
	assert.Len(t, jobs, 1)
	assert.Equal(t, "v2", jobs[0].Command)
}

func TestManager_Delete(t *testing.T) {
	mgr := NewManager()
	err := mgr.Create(Job{Name: "del-me", Schedule: "0 * * * *", Command: "echo"})
	require.NoError(t, err)
	assert.True(t, mgr.Delete("del-me"))
	assert.Empty(t, mgr.List())
}

func TestManager_DeleteMissing(t *testing.T) {
	mgr := NewManager()
	assert.False(t, mgr.Delete("nonexistent"))
}

func TestManager_Get(t *testing.T) {
	mgr := NewManager()
	err := mgr.Create(Job{Name: "get-me", Schedule: "0 * * * *", Command: "echo"})
	require.NoError(t, err)
	job, ok := mgr.Get("get-me")
	assert.True(t, ok)
	assert.Equal(t, "get-me", job.Name)
}

func TestManager_GetMissing(t *testing.T) {
	mgr := NewManager()
	_, ok := mgr.Get("nonexistent")
	assert.False(t, ok)
}

func TestManager_Toggle(t *testing.T) {
	mgr := NewManager()
	err := mgr.Create(Job{Name: "toggle", Schedule: "0 * * * *", Command: "echo", Enabled: true})
	require.NoError(t, err)
	assert.True(t, mgr.Toggle("toggle", false))
	job, _ := mgr.Get("toggle")
	assert.False(t, job.Enabled)
}

func TestManager_ToggleMissing(t *testing.T) {
	mgr := NewManager()
	assert.False(t, mgr.Toggle("nonexistent", true))
}

func TestValidateSchedule(t *testing.T) {
	tests := []struct {
		expr string
		ok   bool
	}{
		{"0 * * * *", true},
		{"*/5 * * * *", true},
		{"0 9 * * 1-5", true},
		{"bad", false},
		{"", false},
		{"60 * * * *", false},
	}
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			err := ValidateSchedule(tt.expr)
			if tt.ok {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
