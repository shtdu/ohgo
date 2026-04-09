package commands

import (
	"context"
	"fmt"
	"strings"
)

type bridgeCmd struct{}

var _ Command = bridgeCmd{}

func (bridgeCmd) Name() string      { return "bridge" }
func (bridgeCmd) ShortHelp() string { return "show bridge subsystem status" }

func (bridgeCmd) Run(ctx context.Context, args string, deps *Deps) (Result, error) {
	if deps.BridgeMgr == nil {
		return Result{Output: "bridge: not available"}, nil
	}

	args = strings.TrimSpace(args)

	// Connect all bridges on demand.
	if args == "connect" {
		if err := deps.BridgeMgr.ConnectAll(ctx); err != nil {
			return Result{Output: fmt.Sprintf("bridge: connect failed: %v", err)}, nil
		}
		return Result{Output: "bridge: all bridges connected"}, nil
	}

	// Show status.
	statuses := deps.BridgeMgr.Status()
	if len(statuses) == 0 {
		return Result{Output: "bridge: no bridges registered"}, nil
	}

	var b strings.Builder
	b.WriteString("Bridge Status:\n")
	for _, s := range statuses {
		state := "disconnected"
		if s.Connected {
			state = "connected"
		}
		fmt.Fprintf(&b, "  %-12s %s\n", s.Name, state)
	}
	return Result{Output: b.String()}, nil
}
