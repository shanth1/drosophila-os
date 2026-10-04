package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shanth1/drosophila-os/internal/engine"
	"github.com/shanth1/drosophila-os/internal/wasm"
	"github.com/shanth1/drosophila-os/plugins"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg, err := parseConfig(os.Args[1:], os.Stdout)
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(2)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx, cfg, logger); err != nil {
		logger.Error("application failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg Config, logger *slog.Logger) error {
	if cfg.Mode != "test" && cfg.Mode != "brain" {
		return fmt.Errorf("mode must be test or brain")
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

func runBrain(ctx context.Context, cfg Config, logger *slog.Logger) (err error) {
	logger.Info("application starting", "brain", cfg.BrainPath)

	eng, err := engine.LoadEngine(cfg.BrainPath)
	if err != nil {
		return err
	}
	if eng.Conn.NumNeurons <= 100 {
		return fmt.Errorf("random sensor requires neuron index 100; graph has %d neurons", eng.Conn.NumNeurons)
	}
	logger.Info("brain loaded", "neurons", eng.Conn.NumNeurons, "synapses", eng.Conn.NumEdges)

	wasmManager, err := wasm.NewManager(ctx, eng.Buffer)
	if err != nil {
		return err
	}
	tickNum := 0
	var totalSpikes uint64
	defer func() {
		if closeErr := wasmManager.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close WASM runtime: %w", closeErr))
		}
		// Polling has stopped; include failures from the last in-flight call.
		select {
		case sensorErr := <-wasmManager.Errors():
			err = errors.Join(err, sensorErr)
		default:
		}
		if err == nil {
			reason := "tick limit"
			if ctx.Err() != nil {
				reason = "context canceled"
			}
			logger.Info("application stopped", "reason", reason, "ticks", tickNum, "total_spikes", totalSpikes)
		}
	}()

	// Load the embedded random sensor (polls every 50ms).
	err = wasmManager.LoadSensor("sensor_random", plugins.SensorRandomWASM, 50*time.Millisecond)
	if err != nil {
		return err
	}
	logger.Info("sensor loaded", "sensor", "sensor_random", "interval", 50*time.Millisecond)

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	logger.Info("engine started", "tick_interval", 10*time.Millisecond, "target_neuron", 100, "tick_limit", cfg.TickLimit)
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-wasmManager.Errors():
			return err
		case <-ticker.C:
		}
		tickNum++
		eng.Tick()

		// Count spikes in this tick
		spikes := 0
		for _, fired := range eng.State.Fired {
			if fired {
				spikes++
			}
		}
		totalSpikes += uint64(spikes)

		if tickNum%50 == 0 {
			// Print every 50 ticks (nominally 500ms).
			logger.Info("engine state", "tick", tickNum, "spikes", spikes,
				"total_spikes", totalSpikes, "target_voltage", eng.State.Voltages[100])
		}

		if cfg.TickLimit > 0 && tickNum >= cfg.TickLimit {
			break
		}
	}

	return nil
}
