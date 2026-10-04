// Package app composes the Drosophila runtime independently of its CLI.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
)

// Run serves the embedded UI alongside the selected backend until cancellation,
// an explicit tick limit, or a component failure. It waits for both to stop.
func Run(ctx context.Context, cfg Config, logger *slog.Logger) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return nil
	}
	// Bind before starting the backend, so an occupied address fails immediately.
	listener, err := net.Listen("tcp", cfg.ListenAddress)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	return runWithListener(ctx, cfg, logger, listener)
}

// runWithListener owns both components and waits for both to finish before returning.
func runWithListener(ctx context.Context, cfg Config, logger *slog.Logger, listener net.Listener) error {
	defer listener.Close()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	results := make(chan error, 2)
	go func() { results <- serveUI(ctx, listener, logger) }()
	go func() { results <- runBackend(ctx, cfg, logger) }()
	first := <-results
	// Failure, an explicit tick limit, or Ctrl+C ends the entire application.
	cancel()
	second := <-results
	return errors.Join(first, second)
}

func runBackend(ctx context.Context, cfg Config, logger *slog.Logger) error {
	switch cfg.Mode {
	case "test":
		return runTestRuntime(ctx, cfg, logger)
	case "brain":
		return runBrain(ctx, cfg, logger)
	default:
		return fmt.Errorf("mode must be test or brain")
	}
}
