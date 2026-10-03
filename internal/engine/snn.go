package engine

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
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
	if _, err := io.ReadFull(f, magic); err != nil {
		return nil, fmt.Errorf("failed to read magic bytes: %w", err)
	}
	if string(magic) != "DROS" {
		return nil, fmt.Errorf("invalid magic bytes, expected DROS")
	}

	var version, numNeurons, numEdges uint32
	for _, field := range []*uint32{&version, &numNeurons, &numEdges} {
		if err := binary.Read(f, order, field); err != nil {
			return nil, fmt.Errorf("failed to read connectome header: %w", err)
		}
	}
	if version != 1 {
		return nil, fmt.Errorf("unsupported connectome version: %d", version)
	}
	if numNeurons == 0 {
		return nil, fmt.Errorf("connectome must contain neurons")
	}
	offsetCount := uint64(numNeurons) + 1
	expectedSize := uint64(16) + 4*offsetCount + 8*uint64(numEdges)
	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat connectome: %w", err)
	}
	if info.Size() < 0 || uint64(info.Size()) != expectedSize {
		return nil, fmt.Errorf("invalid connectome size: got %d, expected %d", info.Size(), expectedSize)
	}
	maxInt := uint64(^uint(0) >> 1)
	if offsetCount > maxInt/4 || uint64(numEdges) > maxInt/4 {
		return nil, fmt.Errorf("connectome arrays exceed platform limits")
	}

	conn := &Connectome{
		NumNeurons: numNeurons,
		NumEdges:   numEdges,
		Offsets:    make([]uint32, int(offsetCount)),
		Targets:    make([]uint32, numEdges),
		Weights:    make([]float32, numEdges),
	}

	for _, data := range []any{conn.Offsets, conn.Targets, conn.Weights} {
		if err := binary.Read(f, order, data); err != nil {
			return nil, fmt.Errorf("failed to read connectome arrays: %w", err)
		}
	}
	if conn.Offsets[0] != 0 || conn.Offsets[len(conn.Offsets)-1] != numEdges {
		return nil, fmt.Errorf("invalid CSR offset endpoints")
	}
	for i := 1; i < len(conn.Offsets); i++ {
		if conn.Offsets[i] < conn.Offsets[i-1] {
			return nil, fmt.Errorf("CSR offsets must be monotonic")
		}
	}
	for i, target := range conn.Targets {
		if target >= numNeurons {
			return nil, fmt.Errorf("edge %d target out of bounds: %d", i, target)
		}
		weight := float64(conn.Weights[i])
		if math.IsNaN(weight) || math.IsInf(weight, 0) {
			return nil, fmt.Errorf("edge %d weight must be finite", i)
		}
	}

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
