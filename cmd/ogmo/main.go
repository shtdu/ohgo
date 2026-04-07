// Package main is the entrypoint for the ogmo personal agent binary.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/shtdu/ohgo/internal/channels"
	"github.com/shtdu/ohgo/internal/config"
	"github.com/shtdu/ohgo/internal/engine"
)

func main() {
	var (
		configDir string
		channel   string
	)

	flag.StringVar(&configDir, "config", "", "config directory path")
	flag.StringVar(&channel, "channel", "", "IM channel to connect (telegram, slack, discord, feishu)")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Load config
	cfgMgr := config.NewManager(configDir)
	cfg, err := cfgMgr.Load(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create engine
	eng := engine.New(engine.Options{
		Model: cfg.Model,
	})

	// If a channel is specified, start in channel mode
	if channel != "" {
		_ = channels.NewRegistry() // TODO: register channel implementations
		fmt.Printf("ogmo - connecting to %s (not yet implemented)\n", channel)
		return
	}

	// Default: run as personal agent
	_ = eng
	fmt.Println("ogmo - Personal Agent")
	fmt.Println("Not yet implemented.")
}
