package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
