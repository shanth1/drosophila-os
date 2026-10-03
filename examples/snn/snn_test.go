package snn_test

import (
	"testing"

	"github.com/shanth1/drosophila-os/internal/engine"
)

func TestSpikePropagation(t *testing.T) {
	// A is neuron 0, B is neuron 1. The only connection is A -> B, weight 1.
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
			DecayRates:      []float32{0, 0}, // Disable leak to isolate propagation.
			Fired:           []bool{false, false},
		},
		Buffer: engine.NewInputBuffer(2),
	}

	// An external signal stays in the input buffer until Tick consumes it.
	e.Buffer.EmitSignal(0, 1)
	t.Logf("Before tick: voltage=%v pending=%v fired=%v",
		e.State.Voltages, e.State.PendingVoltages, e.State.Fired)
	if e.State.Voltages[0] != 0 || e.State.PendingVoltages[0] != 0 || e.State.Fired[0] {
		t.Fatal("buffering an input must not change neuron A's state yet")
	}

	// A fires and resets. Its outgoing signal is queued for B's next tick.
	e.Tick()
	t.Logf("After tick 1: voltage=%v pending=%v fired=%v",
		e.State.Voltages, e.State.PendingVoltages, e.State.Fired)
	if !e.State.Fired[0] || e.State.Fired[1] {
		t.Fatal("only A must fire on tick 1")
	}
	if e.State.Voltages[0] != 0 || e.State.Voltages[1] != 0 ||
		e.State.PendingVoltages[0] != 0 || e.State.PendingVoltages[1] != 1 {
		t.Fatal("A must reset and queue weight 1 for B without changing B's voltage")
	}

	// No new external input: B consumes the queued signal and fires.
	e.Tick()
	t.Logf("After tick 2: voltage=%v pending=%v fired=%v",
		e.State.Voltages, e.State.PendingVoltages, e.State.Fired)
	if e.State.Fired[0] || !e.State.Fired[1] {
		t.Fatal("only B must fire on tick 2")
	}
	if e.State.Voltages[0] != 0 || e.State.Voltages[1] != 0 ||
		e.State.PendingVoltages[0] != 0 || e.State.PendingVoltages[1] != 0 {
		t.Fatal("both voltages and pending inputs must be zero after tick 2")
	}
}
