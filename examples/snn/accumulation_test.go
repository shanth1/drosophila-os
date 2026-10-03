package snn_test

import (
	"testing"

	"github.com/shanth1/drosophila-os/internal/engine"
)

func TestVoltageAccumulationAndLeak(t *testing.T) {
	// One neuron, no connections: only input, leak, threshold, and reset.
	e := &engine.Engine{
		Conn: &engine.Connectome{
			NumNeurons: 1,
			Offsets:    []uint32{0, 0},
		},
		State: &engine.State{
			Voltages:        []float32{0},
			PendingVoltages: []float32{0},
			Thresholds:      []float32{1},
			DecayRates:      []float32{0.5},
			Fired:           []bool{false},
		},
		Buffer: engine.NewInputBuffer(1),
	}

	// Consecutive ticks of the same neuron, not independent test cases.
	steps := []struct {
		input       float32
		wantVoltage float32
		wantFired   bool
	}{
		{input: 0.5, wantVoltage: 0.5},                 // 0 * 0.5 + 0.5
		{input: 0, wantVoltage: 0.25},                  // 0.5 * 0.5, no new input
		{input: 0.5, wantVoltage: 0.625},               // 0.25 * 0.5 + 0.5
		{input: 0.75, wantVoltage: 0, wantFired: true}, // 0.625 * 0.5 + 0.75 = 1.0625; fire and reset
	}
	for i, step := range steps {
		if step.input != 0 {
			e.Buffer.EmitSignal(0, step.input)
		}
		e.Tick()
		t.Logf("Tick %d: input=%v voltage=%v fired=%v",
			i+1, step.input, e.State.Voltages[0], e.State.Fired[0])
		if e.State.Voltages[0] != step.wantVoltage || e.State.Fired[0] != step.wantFired {
			t.Fatalf("tick %d: expected voltage=%v fired=%v, got voltage=%v fired=%v",
				i+1, step.wantVoltage, step.wantFired, e.State.Voltages[0], e.State.Fired[0])
		}
		if e.State.PendingVoltages[0] != 0 {
			t.Fatalf("tick %d: pending input must be consumed", i+1)
		}
	}
}
