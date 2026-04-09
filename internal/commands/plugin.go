package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/shtdu/ohgo/internal/plugins"
)

type pluginCmd struct{}

var _ Command = pluginCmd{}

func (pluginCmd) Name() string      { return "plugin" }
func (pluginCmd) ShortHelp() string { return "manage plugins" }

func (pluginCmd) Run(_ context.Context, args string, deps *Deps) (Result, error) {
	if deps.Plugins == nil {
		return Result{Output: "plugin: no plugin manager"}, nil
	}

	args = strings.TrimSpace(args)

	// Handle subcommands.
	if strings.HasPrefix(args, "install ") {
		source := strings.TrimSpace(strings.TrimPrefix(args, "install "))
		if source == "" {
			return Result{Output: "plugin install: missing source path"}, nil
		}
		_, err := plugins.Install(source)
		if err != nil {
			return Result{Output: fmt.Sprintf("plugin install: %v", err)}, nil
		}
		return Result{Output: "plugin: installed from " + source}, nil
	}

	if strings.HasPrefix(args, "remove ") {
		name := strings.TrimSpace(strings.TrimPrefix(args, "remove "))
		if name == "" {
			return Result{Output: "plugin remove: missing plugin name"}, nil
		}
		_, err := plugins.Uninstall(name)
		if err != nil {
			return Result{Output: fmt.Sprintf("plugin remove: %v", err)}, nil
		}
		return Result{Output: "plugin: removed " + name}, nil
	}

	// Default: list plugins.
	pluginList := deps.Plugins.List()
	if len(pluginList) == 0 {
		return Result{Output: "plugin: no plugins loaded"}, nil
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Plugins (%d):\n", len(pluginList))
	for _, p := range pluginList {
		status := "disabled"
		if p.Enabled {
			status = "enabled"
		}
		fmt.Fprintf(&b, "  %s v%s (%s) [%s]", p.Manifest.Name, p.Manifest.Version, p.Path, status)
		if p.Manifest.Description != "" {
			fmt.Fprintf(&b, " - %s", p.Manifest.Description)
		}
		fmt.Fprintln(&b)
	}
	return Result{Output: b.String()}, nil
}
