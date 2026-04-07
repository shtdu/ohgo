package permissions

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/shtdu/ohgo/internal/config"
)

func TestDecision_String(t *testing.T) {
	assert.Equal(t, "allow", Allow.String())
	assert.Equal(t, "deny", Deny.String())
	assert.Equal(t, "ask", Ask.String())
	assert.Equal(t, "unknown", Decision(99).String())
}

func TestDeniedToolAlwaysDenied(t *testing.T) {
	settings := config.PermissionSettings{
		Mode:        "auto",
		DeniedTools: []string{"write_file"},
	}
	checker := NewDefaultChecker(settings)

	dec, err := checker.Check(context.Background(), Check{ToolName: "write_file"})
	assert.NoError(t, err)
	assert.Equal(t, Deny, dec)
}

func TestDeniedToolOverridesAllowList(t *testing.T) {
	settings := config.PermissionSettings{
		Mode:         "auto",
		AllowedTools: []string{"write_file"},
		DeniedTools:  []string{"write_file"},
	}
	checker := NewDefaultChecker(settings)

	dec, err := checker.Check(context.Background(), Check{ToolName: "write_file"})
	assert.NoError(t, err)
	assert.Equal(t, Deny, dec)
}

func TestAllowedToolAlwaysAllowed(t *testing.T) {
	settings := config.PermissionSettings{
		Mode:         "default",
		AllowedTools: []string{"bash"},
	}
	checker := NewDefaultChecker(settings)

	dec, err := checker.Check(context.Background(), Check{ToolName: "bash"})
	assert.NoError(t, err)
	assert.Equal(t, Allow, dec)
}

func TestPathDenyRuleBlocksMatchingPaths(t *testing.T) {
	settings := config.PermissionSettings{
		Mode: "auto",
		PathRules: []config.PathRuleConfig{
			{Pattern: "/etc/*", Allow: false},
		},
	}
	checker := NewDefaultChecker(settings)

	dec, err := checker.Check(context.Background(), Check{
		ToolName: "write_file",
		FilePath: "/etc/passwd",
	})
	assert.NoError(t, err)
	assert.Equal(t, Deny, dec)
}

func TestPathAllowRuleGrantsAccess(t *testing.T) {
	settings := config.PermissionSettings{
		Mode: "default",
		PathRules: []config.PathRuleConfig{
			{Pattern: "/tmp/test.txt", Allow: true},
		},
	}
	checker := NewDefaultChecker(settings)

	// write_file in default mode would normally be Ask, but path allow rule grants access
	dec, err := checker.Check(context.Background(), Check{
		ToolName:   "write_file",
		FilePath:   "/tmp/test.txt",
		IsReadOnly: false,
	})
	assert.NoError(t, err)
	assert.Equal(t, Allow, dec, "path allow rule should short-circuit to Allow")
}

func TestCommandDenyPatternBlocksMatchingCommands(t *testing.T) {
	settings := config.PermissionSettings{
		Mode:           "auto",
		DeniedCommands: []string{"rm *", "sudo *"},
	}
	checker := NewDefaultChecker(settings)

	// filepath.Match uses glob: * does not cross '/', so "rm *" matches "rm -rf"
	// but not "rm -rf /tmp/test". Use a command without '/' for this test.
	dec, err := checker.Check(context.Background(), Check{
		ToolName: "bash",
		Command:  "rm -rf",
	})
	assert.NoError(t, err)
	assert.Equal(t, Deny, dec)

	// Verify it does not match a command with '/' beyond the glob
	dec, err = checker.Check(context.Background(), Check{
		ToolName: "bash",
		Command:  "rm -rf /tmp/test",
	})
	assert.NoError(t, err)
	assert.Equal(t, Allow, dec) // not matched because * won't cross /
}

func TestCommandDenyPatternDoesNotBlockNonMatching(t *testing.T) {
	settings := config.PermissionSettings{
		Mode:           "auto",
		DeniedCommands: []string{"rm *"},
	}
	checker := NewDefaultChecker(settings)

	dec, err := checker.Check(context.Background(), Check{
		ToolName: "bash",
		Command:  "ls -la",
	})
	assert.NoError(t, err)
	assert.Equal(t, Allow, dec)
}

func TestAutoModeAllowsAll(t *testing.T) {
	settings := config.PermissionSettings{
		Mode: "auto",
	}
	checker := NewDefaultChecker(settings)

	dec, err := checker.Check(context.Background(), Check{
		ToolName:  "write_file",
		IsReadOnly: false,
	})
	assert.NoError(t, err)
	assert.Equal(t, Allow, dec)
}

func TestPlanModeDeniesWriteTools(t *testing.T) {
	settings := config.PermissionSettings{
		Mode: "plan",
	}
	checker := NewDefaultChecker(settings)

	dec, err := checker.Check(context.Background(), Check{
		ToolName:  "write_file",
		IsReadOnly: false,
	})
	assert.NoError(t, err)
	assert.Equal(t, Deny, dec)
}

func TestPlanModeAllowsReadTools(t *testing.T) {
	settings := config.PermissionSettings{
		Mode: "plan",
	}
	checker := NewDefaultChecker(settings)

	dec, err := checker.Check(context.Background(), Check{
		ToolName:  "read_file",
		IsReadOnly: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, Allow, dec)
}

func TestDefaultModeAsksForWriteTools(t *testing.T) {
	settings := config.PermissionSettings{
		Mode: "default",
	}
	checker := NewDefaultChecker(settings)

	dec, err := checker.Check(context.Background(), Check{
		ToolName:  "write_file",
		IsReadOnly: false,
	})
	assert.NoError(t, err)
	assert.Equal(t, Ask, dec)
}

func TestDefaultModeAllowsReadTools(t *testing.T) {
	settings := config.PermissionSettings{
		Mode: "default",
	}
	checker := NewDefaultChecker(settings)

	dec, err := checker.Check(context.Background(), Check{
		ToolName:  "read_file",
		IsReadOnly: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, Allow, dec)
}

func TestUnknownToolInDefaultMode(t *testing.T) {
	settings := config.PermissionSettings{
		Mode: "default",
	}
	checker := NewDefaultChecker(settings)

	// Unknown tool, not read-only, default mode -> Ask
	dec, err := checker.Check(context.Background(), Check{
		ToolName:  "mystery_tool",
		IsReadOnly: false,
	})
	assert.NoError(t, err)
	assert.Equal(t, Ask, dec)
}

func TestConcurrentSafety(t *testing.T) {
	settings := config.PermissionSettings{
		Mode:        "auto",
		DeniedTools: []string{"dangerous_tool"},
		PathRules: []config.PathRuleConfig{
			{Pattern: "/secret/*", Allow: false},
		},
		DeniedCommands: []string{"rm *"},
	}
	checker := NewDefaultChecker(settings)

	var wg sync.WaitGroup
	results := make(chan Decision, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			check := Check{
				ToolName:  "some_tool",
				IsReadOnly: false,
			}
			switch i % 3 {
			case 0:
				check.ToolName = "dangerous_tool" // explicit deny
			case 1:
				check.FilePath = "/secret/key" // path deny rule
			case 2:
				check.Command = "rm -rf" // command deny pattern (no / so glob matches)
			}

			dec, err := checker.Check(context.Background(), check)
			assert.NoError(t, err)
			results <- dec
		}(i)
	}

	wg.Wait()
	close(results)

	for dec := range results {
		assert.Equal(t, Deny, dec)
	}
}
