// Package app composes the Drosophila runtime independently of its CLI.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/shanth1/drosophila-os/internal/hostapi"
	"github.com/shanth1/drosophila-os/internal/telemetry"
	"github.com/shanth1/drosophila-os/internal/testenv"
	"github.com/shanth1/drosophila-os/ui"
)

// Run serves the embedded UI alongside the selected backend until cancellation,
// an explicit tick limit, or a component failure. It waits for all components to stop.
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

// runWithListener owns HTTP, backend, and optional test environment lifetimes.
func runWithListener(ctx context.Context, cfg Config, logger *slog.Logger, listener net.Listener) error {
	defer listener.Close()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	hub, err := telemetry.New(telemetry.State{Mode: cfg.Mode, Status: "starting"})
	if err != nil {
		return err
	}
	defer hub.Close()
	api := hostapi.New(hub)
	defer api.Close()
	mux := http.NewServeMux()
	mux.Handle("/api/v1/", api)
	mux.Handle("/", ui.Handler())
	results := make(chan error, 3)
	components := 2
	var targetURL string
	if cfg.Mode == "test" {
		testListener, err := net.Listen("tcp", cfg.TestListenAddress)
		if err != nil {
			return fmt.Errorf("test environment listen: %w", err)
		}
		defer testListener.Close()
		address := testListener.Addr().(*net.TCPAddr)
		host := address.IP.String()
		if address.IP.IsUnspecified() {
			if address.IP.To4() != nil {
				host = "127.0.0.1"
			} else {
				host = "::1"
			}
		}
		targetURL = "http://" + net.JoinHostPort(host, fmt.Sprint(address.Port)) + "/health"
		components++
		go func() { results <- serveHTTP(ctx, testListener, testenv.New(), logger, "test environment listening") }()
	}
	go func() { results <- serveHTTP(ctx, listener, mux, logger, "frontend listening") }()
	go func() { results <- runBackend(ctx, cfg, logger, targetURL, hub) }()
	first := <-results
	// Failure, an explicit tick limit, or Ctrl+C ends the entire application.
	cancel()
	for remaining := components - 1; remaining > 0; remaining-- {
		first = errors.Join(first, <-results)
	}
	return first
}

func runBackend(ctx context.Context, cfg Config, logger *slog.Logger, targetURL string, hub *telemetry.Hub) error {
	switch cfg.Mode {
	case "test":
		return runTestRuntime(ctx, cfg, logger, targetURL, hub)
	case "brain":
		return runBrain(ctx, cfg, logger, hub)
	default:
		return fmt.Errorf("mode must be test or brain")
	}
}
