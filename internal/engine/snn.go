package engine

import (
	"encoding/binary"
	"fmt"
	"os"
)

// Connectome holds the static biological wiring in CSR (Compressed Sparse Row) format.
// This is read-only during runtime.
type Connectome struct {
	NumNeurons uint32
	NumEdges   uint32
	Offsets    []uint32
	Targets    []uint32
	Weights    []float32
}

// State holds the dynamic, mutable state of the SNN.
// Designed as Struct-of-Arrays (SoA) for maximum CPU cache efficiency.
type State struct {
	Voltages        []float32 // Current electrical potential
	PendingVoltages []float32 // Buffer for incoming spikes (ensures tick determinism)
	Thresholds      []float32 // Action potential threshold
	DecayRates      []float32 // Habituation/Leak rate
	Fired           []bool    // Track which neurons spiked this tick (for Trace Logs/UI)
}

// Engine is the core SNN runtime.
type Engine struct {
	Conn   *Connectome
	State  *State
	Buffer *InputBuffer
}

// LoadEngine initializes the brain from a compiled binary connectome.
func LoadEngine(binPath string) (*Engine, error) {
	f, err := os.Open(binPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open connectome: %w", err)
	}
	defer f.Close()

	order := binary.LittleEndian
	magic := make([]byte, 4)
	if _, err := f.Read(magic); err != nil || string(magic) != "DROS" {
		return nil, fmt.Errorf("invalid magic bytes, expected DROS")
	}

	var version, numNeurons, numEdges uint32
	binary.Read(f, order, &version)
	binary.Read(f, order, &numNeurons)
	binary.Read(f, order, &numEdges)

	conn := &Connectome{
		NumNeurons: numNeurons,
		NumEdges:   numEdges,
		Offsets:    make([]uint32, numNeurons+1),
		Targets:    make([]uint32, numEdges),
		Weights:    make([]float32, numEdges),
	}

	binary.Read(f, order, conn.Offsets)
	binary.Read(f, order, conn.Targets)
	binary.Read(f, order, conn.Weights)

	// Initialize dynamic state
	state := &State{
		Voltages:        make([]float32, numNeurons),
		PendingVoltages: make([]float32, numNeurons),
		Thresholds:      make([]float32, numNeurons),
		DecayRates:      make([]float32, numNeurons),
		Fired:           make([]bool, numNeurons),
	}

	// Set default biology rules
	for i := uint32(0); i < numNeurons; i++ {
		state.Thresholds[i] = 1.0  // Default spike threshold
		state.DecayRates[i] = 0.05 // Default leak (5% per tick)
	}

	return &Engine{
		Conn:   conn,
		State:  state,
		Buffer: NewInputBuffer(numNeurons),
	}, nil
}
