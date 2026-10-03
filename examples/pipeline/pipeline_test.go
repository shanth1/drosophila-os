package pipeline_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/shanth1/drosophila-os/internal/engine"
	"github.com/shanth1/drosophila-os/internal/wasm"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func TestWASMSignalTriggersOutput(t *testing.T) {
	ctx := context.Background()

	// Build the guest sensor into a temporary file, not a production asset.
	wasmPath := filepath.Join(t.TempDir(), "sensor.wasm")
	build := exec.CommandContext(ctx, "go", "build", "-mod=readonly",
		"-buildmode=c-shared", "-o", wasmPath, "./testdata/sensor")
	build.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm", "CGO_ENABLED=0")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build guest: %v\n%s", err, output)
	}
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatal(err)
	}

	// A (index 0) accumulates input; its only connection targets B (index 1).
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

	// The production ABI delivers guest signals directly to the engine buffer.
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, runtime); err != nil {
		t.Fatal(err)
	}
	if err := wasm.RegisterInputABI(ctx, runtime, e.Buffer); err != nil {
		t.Fatal(err)
	}
	guest, err := runtime.InstantiateWithConfig(ctx, wasmBytes,
		wazero.NewModuleConfig().WithStartFunctions("_initialize"))
	if err != nil {
		t.Fatal(err)
	}
	sensorTick := guest.ExportedFunction("tick")
	if sensorTick == nil {
		t.Fatal("guest is missing exported function tick")
	}

	// These are consecutive steps of one network, not independent test cases.
	steps := []struct {
		pollSensor bool
		voltage    [2]float32
		pending    [2]float32
		fired      [2]bool
	}{
		{pollSensor: true, voltage: [2]float32{0.5, 0}},
		{pollSensor: true, pending: [2]float32{0, 1}, fired: [2]bool{true, false}},
		{pollSensor: false, fired: [2]bool{false, true}},
	}
	var eventTicks []int
	for i, step := range steps {
		if step.pollSensor {
			if _, err := sensorTick.Call(ctx); err != nil {
				t.Fatal(err)
			}
		}
		e.Tick()
		t.Logf("Tick %d: sensor=%v voltage=%v pending=%v fired=%v",
			i+1, step.pollSensor, e.State.Voltages, e.State.PendingVoltages, e.State.Fired)
		for neuron := 0; neuron < 2; neuron++ {
			if e.State.Voltages[neuron] != step.voltage[neuron] ||
				e.State.PendingVoltages[neuron] != step.pending[neuron] ||
				e.State.Fired[neuron] != step.fired[neuron] {
				t.Fatalf("tick %d neuron %d: expected voltage=%v pending=%v fired=%v",
					i+1, neuron, step.voltage[neuron], step.pending[neuron], step.fired[neuron])
			}
		}
		// Host output: record an event when B fires; no WASM effector is involved.
		if e.State.Fired[1] {
			eventTicks = append(eventTicks, i+1)
			t.Logf("OUTPUT: neuron B fired on tick %d", i+1)
		}
	}
	if len(eventTicks) != 1 || eventTicks[0] != 3 {
		t.Fatalf("expected exactly one output event on tick 3, got %v", eventTicks)
	}
}
