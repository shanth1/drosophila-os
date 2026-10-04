package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shanth1/drosophila-os/internal/engine"
	"github.com/shanth1/drosophila-os/internal/telemetry"
	"github.com/shanth1/drosophila-os/internal/wasm"
	"github.com/shanth1/drosophila-os/plugins"
)

func runBrain(ctx context.Context, cfg Config, logger *slog.Logger, hub *telemetry.Hub) (err error) {
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
	state := telemetry.State{
		Mode: "brain", Status: "running",
		Brain:    telemetry.Brain{Neurons: eng.Conn.NumNeurons, Synapses: eng.Conn.NumEdges},
		Modules:  []telemetry.Module{{ID: "sensor_random", Kind: "sensor", Status: "running", Capabilities: []string{"signal.normalized.v1"}}},
		Entities: []telemetry.Entity{{ID: "brain-network", Kind: "neural-network", Label: "Experimental biological graph"}},
	}
	if err := hub.Publish(state, nil); err != nil {
		return err
	}

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
		// Sample overview telemetry at 10 Hz, not on every biological tick.
		if tickNum%10 == 0 || (cfg.TickLimit > 0 && tickNum == cfg.TickLimit) {
			state.Brain.Tick, state.Brain.Spikes, state.Brain.TotalSpikes = uint64(tickNum), uint32(spikes), totalSpikes
			if err := hub.Publish(state, nil); err != nil {
				return err
			}
		}

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
