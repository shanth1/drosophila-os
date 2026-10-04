package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shanth1/drosophila-os/internal/engine"
	"github.com/shanth1/drosophila-os/internal/httpmonitor"
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

func runTestRuntime(ctx context.Context, cfg Config, logger *slog.Logger, targetURL string) (err error) {
	e := newTestEngine()
	probe, err := httpmonitor.New(targetURL, 2*time.Second)
	if err != nil {
		return err
	}
	defer probe.Close()
	var lastStatus uint32
	var lastLog time.Time
	manager, err := wasm.NewManagerWithHTTPProbe(ctx, e.Buffer, func(ctx context.Context) (uint32, uint32) {
		result := probe.Measure(ctx)
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
		logger.Debug("test runtime state", "tick", tickNum,
			"voltage", e.State.Voltages, "pending", e.State.PendingVoltages, "fired", e.State.Fired)
		if e.State.Fired[2] {
			events++
			logger.Info("output event", "tick", tickNum, "neuron", 2)
		}
	}
	return nil
}
