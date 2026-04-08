package skills

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFrontmatter_YAMLWithNameAndDescription(t *testing.T) {
	content := `---
name: commit
description: Create a git commit
---
## Steps
1. Stage files
2. Write message
`
	name, desc, body := parseFrontmatter("default", content)
	assert.Equal(t, "commit", name)
	assert.Equal(t, "Create a git commit", desc)
	assert.Contains(t, body, "## Steps")
}

func TestParseFrontmatter_YAMLWithQuotedValues(t *testing.T) {
	content := `---
name: "review-pr"
description: 'Review a pull request'
---
Body text here.
`
	name, desc, body := parseFrontmatter("default", content)
	assert.Equal(t, "review-pr", name)
	assert.Equal(t, "Review a pull request", desc)
	assert.Contains(t, body, "Body text here.")
}

func TestParseFrontmatter_NoFrontmatterWithHeading(t *testing.T) {
	content := `# My Skill

This is the first paragraph that describes the skill in detail.
More content here.
`
	name, desc, _ := parseFrontmatter("default", content)
	assert.Equal(t, "My Skill", name)
	assert.Equal(t, "This is the first paragraph that describes the skill in detail.", desc)
}

func TestParseFrontmatter_NoFrontmatterNoHeading(t *testing.T) {
	content := `Just a paragraph without any heading or frontmatter.
`
	name, desc, _ := parseFrontmatter("fallback-name", content)
	assert.Equal(t, "fallback-name", name)
	assert.Equal(t, "Just a paragraph without any heading or frontmatter.", desc)
}

func TestParseFrontmatter_EmptyContent(t *testing.T) {
	name, desc, body := parseFrontmatter("empty", "")
	assert.Equal(t, "empty", name)
	assert.Equal(t, "Skill: empty", desc)
	assert.Equal(t, "", body)
}

func TestParseFrontmatter_OnlyFrontmatterNoBody(t *testing.T) {
	content := `---
name: solo
description: Just frontmatter nothing else
---
`
	name, desc, body := parseFrontmatter("default", content)
	assert.Equal(t, "solo", name)
	assert.Equal(t, "Just frontmatter nothing else", desc)
	assert.Equal(t, "", body)
}

func TestParseFrontmatter_DescriptionCappedAt200(t *testing.T) {
	longDesc := ""
	for i := 0; i < 300; i++ {
		longDesc += "x"
	}
	content := longDesc + "\n"

	name, desc, _ := parseFrontmatter("default", content)
	assert.Equal(t, "default", name)
	assert.Len(t, desc, 200)
}

func TestParseFrontmatter_YAMLWithEmptyName(t *testing.T) {
	content := `---
name:
description: Has desc but no name
---
Body.
`
	name, desc, _ := parseFrontmatter("fallback", content)
	assert.Equal(t, "fallback", name)
	assert.Equal(t, "Has desc but no name", desc)
}

func TestParseFrontmatter_YAMLWithEmptyDescription(t *testing.T) {
	content := `---
name: named
description:
---
Some body text.
`
	name, desc, body := parseFrontmatter("default", content)
	assert.Equal(t, "named", name)
	// Empty description falls back to first body paragraph after frontmatter.
	assert.Equal(t, "Some body text.", desc)
	assert.Contains(t, body, "Some body text.")
}
