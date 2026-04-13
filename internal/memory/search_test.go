package memory

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenize_ASCII(t *testing.T) {
	tokens := tokenize("Hello World testing search")
	assert.True(t, tokens["hello"])
	assert.True(t, tokens["world"])
	assert.True(t, tokens["testing"])
	assert.True(t, tokens["search"])
	// Short words should be excluded.
	assert.False(t, tokens["is"])
}

func TestTokenize_Han(t *testing.T) {
	tokens := tokenize("测试搜索功能")
	assert.True(t, tokens["测"])
	assert.True(t, tokens["搜"])
}

func TestTokenize_Empty(t *testing.T) {
	tokens := tokenize("")
	assert.Empty(t, tokens)
}

func TestTokenize_Short(t *testing.T) {
	tokens := tokenize("a b c")
	assert.Empty(t, tokens)
}

func TestFind_EmptyQuery(t *testing.T) {
	results, err := Find("", t.TempDir(), 5)
	assert.NoError(t, err)
	assert.Nil(t, results)
}

func TestFind_WithRealFiles(t *testing.T) {
	t.Setenv("OHGO_DATA_DIR", t.TempDir())
	cwd := t.TempDir()
	memDir, err := ProjectDir(cwd)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(filepath.Join(memDir, "golang.md"), []byte("---\nname: Go Testing\ndescription: golang test patterns\n---\nHow to write tests in Go\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(memDir, "python.md"), []byte("---\nname: Python Web\ndescription: django flask\n---\nBuilding web apps\n"), 0o644))

	results, err := Find("golang testing", cwd, 5)
	require.NoError(t, err)
	require.NotEmpty(t, results, "expected at least one result")
	assert.Equal(t, "Go Testing", results[0].Title)
}

func TestFind_MaxResultsDefault(t *testing.T) {
	t.Setenv("OHGO_DATA_DIR", t.TempDir())
	cwd := t.TempDir()
	memDir, err := ProjectDir(cwd)
	require.NoError(t, err)

	for i := 0; i < 6; i++ {
		name := filepath.Join(memDir, fmt.Sprintf("match%d.md", i))
		content := fmt.Sprintf("---\nname: Match%d\ndescription: searchterm item\n---\nsearchterm body %d\n", i, i)
		require.NoError(t, os.WriteFile(name, []byte(content), 0o644))
	}

	results, err := Find("searchterm", cwd, 0)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(results), 5)
}

func TestFind_ScoreOrdering(t *testing.T) {
	t.Setenv("OHGO_DATA_DIR", t.TempDir())
	cwd := t.TempDir()
	memDir, err := ProjectDir(cwd)
	require.NoError(t, err)

	// high.md: "keyword" appears in name AND description => 2 metaHits => score 4
	require.NoError(t, os.WriteFile(filepath.Join(memDir, "high.md"), []byte("---\nname: keyword match\ndescription: keyword here\n---\nbody text\n"), 0o644))
	// low.md: no "keyword" in metadata at all, only in body => 1 bodyHit => score 1
	require.NoError(t, os.WriteFile(filepath.Join(memDir, "low.md"), []byte("---\nname: unrelated\ndescription: something else\n---\nkeyword in body only\n"), 0o644))

	results, err := Find("keyword", cwd, 5)
	require.NoError(t, err)
	require.Len(t, results, 2, "expected two results")
	assert.Equal(t, "keyword match", results[0].Title, "higher-scoring result should be first")
}
