package engine

// ApplyHabituation implements the 'Leak' in the Leaky Integrate-and-Fire (LIF) model.
// Voltages decay over time, preventing the system from spiking constantly due to background noise.
func (e *Engine) ApplyHabituation() {
	for i := uint32(0); i < e.Conn.NumNeurons; i++ {
		// If DecayRate is 0.0 (Absolute Nociceptor), it will not habituate.
		// Otherwise, it leaks potential towards 0.0.
		e.State.Voltages[i] -= e.State.Voltages[i] * e.State.DecayRates[i]

		// Prevent negative voltages (hyperpolarization is ignored in this simplified model)
		if e.State.Voltages[i] < 0 {
			e.State.Voltages[i] = 0
		}
	}
}

// MarkAbsoluteNociceptor disables habituation for critical sensors (e.g., Fatal Errors, Memory Leaks).
func (e *Engine) MarkAbsoluteNociceptor(neuronID uint32) {
	if neuronID < e.Conn.NumNeurons {
		e.State.DecayRates[neuronID] = 0.0
	}
}
