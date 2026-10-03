package engine

import (
	"math"
	"sync/atomic"
)

// InputBuffer decouples the asynchronous WASM sensor polling from the Biological Clock.
// Uses atomic operations to avoid mutex locking on the hot path.
type InputBuffer struct {
	// Store float32 as uint32 bits for atomic operations
	signals []uint32
}

func NewInputBuffer(numNeurons uint32) *InputBuffer {
	return &InputBuffer{
		signals: make([]uint32, numNeurons),
	}
}

// EmitSignal is called by the WASM manager when a sensor pushes a normalized value [0.0, 1.0].
func (b *InputBuffer) EmitSignal(neuronID uint32, intensity float32) {
	if neuronID >= uint32(len(b.signals)) {
		return
	}
	bits := math.Float32bits(intensity)
	atomic.StoreUint32(&b.signals[neuronID], bits)
}

// Flush transfers the buffered signals into the brain's pending voltages and resets the buffer.
// Called exclusively by the Biological Clock.
func (b *InputBuffer) Flush(pendingVoltages []float32) {
	for i := 0; i < len(b.signals); i++ {
		bits := atomic.SwapUint32(&b.signals[i], 0) // Read and reset atomically
		if bits != 0 {
			intensity := math.Float32frombits(bits)
			pendingVoltages[i] += intensity
		}
	}
}
