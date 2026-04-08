package skills

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry_IsEmpty(t *testing.T) {
	r := NewRegistry()
	assert.Empty(t, r.List())
	assert.Nil(t, r.Get("anything"))
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	skill := &Skill{Name: "commit", Description: "Create a commit"}
	r.Register(skill)

	got := r.Get("commit")
	require.NotNil(t, got)
	assert.Equal(t, "commit", got.Name)
	assert.Equal(t, "Create a commit", got.Description)
}

func TestRegistry_RegisterOverwrites(t *testing.T) {
	r := NewRegistry()
	r.Register(&Skill{Name: "commit", Description: "v1"})
	r.Register(&Skill{Name: "commit", Description: "v2"})

	got := r.Get("commit")
	require.NotNil(t, got)
	assert.Equal(t, "v2", got.Description)
}

func TestRegistry_GetMissingReturnsNil(t *testing.T) {
	r := NewRegistry()
	r.Register(&Skill{Name: "commit", Description: "exists"})

	assert.Nil(t, r.Get("nonexistent"))
}

func TestRegistry_ListSortedByName(t *testing.T) {
	r := NewRegistry()
	r.Register(&Skill{Name: "zebra"})
	r.Register(&Skill{Name: "alpha"})
	r.Register(&Skill{Name: "middle"})

	list := r.List()
	require.Len(t, list, 3)
	assert.Equal(t, "alpha", list[0].Name)
	assert.Equal(t, "middle", list[1].Name)
	assert.Equal(t, "zebra", list[2].Name)
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := "skill"
			r.Register(&Skill{Name: name, Description: "concurrent"})
			_ = r.Get(name)
		}(i)
	}
	wg.Wait()

	got := r.Get("skill")
	require.NotNil(t, got)
	assert.Equal(t, "concurrent", got.Description)
}
