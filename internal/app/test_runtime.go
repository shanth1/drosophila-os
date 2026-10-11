package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/shanth1/drosophila-os/internal/engine"
	"github.com/shanth1/drosophila-os/internal/httpmonitor"
	"github.com/shanth1/drosophila-os/internal/telemetry"
	"github.com/shanth1/drosophila-os/internal/wasm"
	"github.com/shanth1/drosophila-os/plugins"
)

func newTestEngine() *engine.Engine {
	// Latency (0) and failure (1) feed output (2), not an anomaly classification.
	return &engine.Engine{
		Conn: &engine.Connectome{
			NumNeurons: 3,
			NumEdges:   2,
			Offsets:    []uint32{0, 1, 2, 2},
			Targets:    []uint32{2, 2},
			Weights:    []float32{1, 1},
		},
		State: &engine.State{
			Voltages:        []float32{0, 0, 0},
			PendingVoltages: []float32{0, 0, 0},
			Thresholds:      []float32{1, 1, 1},
			DecayRates:      []float32{0.2, 0.2, 0.2},
			Fired:           []bool{false, false, false},
		},
		Buffer: engine.NewInputBuffer(3),
	}
}

func runTestRuntime(ctx context.Context, cfg Config, logger *slog.Logger, targetURL string, hub *telemetry.Hub) (err error) {
	e := newTestEngine()
	probe, err := httpmonitor.New(targetURL, 2*time.Second)
	if err != nil {
		return err
	}
	defer probe.Close()
	var lastStatus uint32
	var lastLog time.Time
	// Measurements are immutable mailbox values. Only the engine loop publishes state.
	var latest atomic.Pointer[telemetry.Observation]
	manager, err := wasm.NewManagerWithHTTPProbe(ctx, e.Buffer, func(ctx context.Context) (uint32, uint32) {
		result := probe.Measure(ctx)
		measurement := telemetry.HTTPResponse{StatusCode: result.StatusCode, ElapsedMS: result.ElapsedMillis}
		if result.Err != nil {
			measurement.Error = result.Err.Error()
		}
		if ctx.Err() == nil {
			latest.Store(&telemetry.Observation{Source: "sensor_http", Subject: "test-http", Name: "http.response", SchemaVersion: 1, ObservedAt: time.Now().UTC(), Data: measurement})
		}
		if ctx.Err() == nil && (result.StatusCode != lastStatus || time.Since(lastLog) >= time.Second) {
			logger.Info("HTTP observation", "status", result.StatusCode, "elapsed_ms", result.ElapsedMillis, "error", result.Err)
			lastStatus, lastLog = result.StatusCode, time.Now()
		}
		return result.StatusCode, result.ElapsedMillis
	})
	if err != nil {
		return err
	}
	tickNum, events := 0, 0
	defer func() {
		err = errors.Join(err, manager.Close())
		select {
		case sensorErr := <-manager.Errors():
			err = errors.Join(err, sensorErr)
		default:
		}
		if err == nil {
			reason := "tick limit"
			if ctx.Err() != nil {
				reason = "context canceled"
			}
			logger.Info("test runtime stopped", "reason", reason, "ticks", tickNum, "events", events)
		}
	}()
	if err := manager.LoadSensor("sensor_http", plugins.SensorHTTPWASM, 250*time.Millisecond); err != nil {
		return fmt.Errorf("load HTTP sensor: %w", err)
	}
	state := telemetry.State{
		Mode: "test", Status: "running",
		Brain:   telemetry.Brain{Neurons: 3, Synapses: 2},
		Modules: []telemetry.Module{{ID: "sensor_http", Kind: "sensor", Status: "running", Capabilities: []string{"http.response.v1", "signal.normalized.v1"}}},
		Entities: []telemetry.Entity{
			{ID: "test-http", Kind: "service", Label: "Test HTTP service"},
			{ID: "test-network", Kind: "neural-network", Label: "Test network"},
		},
	}
	if err := hub.Publish(state, nil); err != nil {
		return err
	}

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	logger.Info("test runtime started", "tick_interval", 250*time.Millisecond, "tick_limit", cfg.TickLimit,
		"sensor", "sensor_http", "target", targetURL, "neurons", 3, "synapses", 2)
	for cfg.TickLimit == 0 || tickNum < cfg.TickLimit {
		select {
		case <-ctx.Done():
			return nil
		case err := <-manager.Errors():
			return err
		case <-ticker.C:
		}
		tickNum++
		e.Tick()
		state.Brain.Tick = uint64(tickNum)
		state.Brain.Spikes = 0
		for _, fired := range e.State.Fired {
			if fired {
				state.Brain.Spikes++
			}
		}
		state.Brain.TotalSpikes += uint64(state.Brain.Spikes)
		if observation := latest.Load(); observation != nil {
			state.Observations = []telemetry.Observation{*observation}
		}
		logger.Debug("test runtime state", "tick", tickNum,
			"voltage", e.State.Voltages, "pending", e.State.PendingVoltages, "fired", e.State.Fired)
		var event *telemetry.Event
		if e.State.Fired[2] {
			events++
			at := time.Now().UTC()
			state.Brain.OutputEvents = uint64(events)
			state.Brain.LastOutputAt = &at
			event = &telemetry.Event{Source: "snn", Subject: "test-network", Name: "neural.output.spike", SchemaVersion: 1, Data: telemetry.OutputSpike{Tick: uint64(tickNum), Neuron: 2}}
			logger.Info("output event", "tick", tickNum, "neuron", 2)
		}
		if err := hub.Publish(state, event); err != nil {
			return err
		}
	}
	return nil
}
