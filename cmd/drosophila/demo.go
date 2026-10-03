package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/shanth1/drosophila-os/internal/engine"
	"github.com/shanth1/drosophila-os/internal/wasm"
	"github.com/shanth1/drosophila-os/plugins"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func runDemo(ctx context.Context, cfg Config, logger *slog.Logger) (err error) {
	// A (0) accumulates input; B (1) is the output. Leak is disabled.
	e := &engine.Engine{
		Conn: &engine.Connectome{
			NumNeurons: 2,
			NumEdges:   1,
			Offsets:    []uint32{0, 1, 1},
			Targets:    []uint32{1},
			Weights:    []float32{1},
		},
		State: &engine.State{
			Voltages:        []float32{0, 0},
			PendingVoltages: []float32{0, 0},
			Thresholds:      []float32{1, 1},
			DecayRates:      []float32{0, 0},
			Fired:           []bool{false, false},
		},
		Buffer: engine.NewInputBuffer(2),
	}
	runtime := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithCloseOnContextDone(true))
	tickNum, events := 0, 0
	defer func() {
		err = errors.Join(err, runtime.Close(context.Background()))
		if err == nil {
			reason := "tick limit"
			if ctx.Err() != nil {
				reason = "context canceled"
			}
			logger.Info("demo stopped", "reason", reason, "ticks", tickNum, "events", events)
		}
	}()
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, runtime); err != nil {
		return fmt.Errorf("instantiate WASI: %w", err)
	}
	if err := wasm.RegisterInputABI(ctx, runtime, e.Buffer); err != nil {
		return fmt.Errorf("register input ABI: %w", err)
	}
	guest, err := runtime.InstantiateWithConfig(ctx, plugins.SensorFixedWASM,
		wazero.NewModuleConfig().WithName("sensor_fixed").WithStartFunctions("_initialize"))
	if err != nil {
		return fmt.Errorf("load fixed sensor: %w", err)
	}
	sensorTick := guest.ExportedFunction("tick")
	if sensorTick == nil {
		return fmt.Errorf("fixed sensor is missing exported tick")
	}

	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	logger.Info("demo started", "tick_interval", 250*time.Millisecond, "tick_limit", cfg.TickLimit,
		"sensor", "sensor_fixed", "neurons", 2, "synapses", 1)
	for cfg.TickLimit == 0 || tickNum < cfg.TickLimit {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
		tickNum++
		// Repeat the pipeline's two input samples followed by one propagation-only tick.
		pollSensor := tickNum%3 != 0
		if pollSensor {
			if _, err := sensorTick.Call(ctx); err != nil {
				if ctx.Err() != nil {
					return nil
				}
				return fmt.Errorf("fixed sensor tick: %w", err)
			}
		}
		e.Tick()
		logger.Info("demo state", "tick", tickNum, "sensor_polled", pollSensor,
			"voltage", e.State.Voltages, "pending", e.State.PendingVoltages, "fired", e.State.Fired)
		if e.State.Fired[1] {
			events++
			logger.Info("output event", "tick", tickNum, "neuron", 1)
		}
	}
	return nil
}
