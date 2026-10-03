package engine

// Tick represents one cycle of the Biological Clock (e.g., 10ms).
// It executes the LIF math and propagates spikes using the CSR connectome.
// STRICT CONSTRAINT: Zero heap allocations in this function.
func (e *Engine) Tick() {
	numNeurons := e.Conn.NumNeurons
	state := e.State
	conn := e.Conn

	// 1. Flush external sensory inputs into the pending buffer
	e.Buffer.Flush(state.PendingVoltages)

	// 2. Apply Habituation (Leak)
	e.ApplyHabituation()

	// 3. Integrate pending voltages into current state and reset fired tracking
	for i := uint32(0); i < numNeurons; i++ {
		state.Voltages[i] += state.PendingVoltages[i]
		state.PendingVoltages[i] = 0.0
		state.Fired[i] = false
	}

	// 4. Spike Evaluation & Propagation
	for i := uint32(0); i < numNeurons; i++ {
		if state.Voltages[i] >= state.Thresholds[i] {
			// Neuron spikes!
			state.Fired[i] = true
			state.Voltages[i] = 0.0 // Reset potential (refractory period simplified)

			// Fast traversal of target synapses using CSR Offsets
			start := conn.Offsets[i]
			end := conn.Offsets[i+1]

			for j := start; j < end; j++ {
				target := conn.Targets[j]
				weight := conn.Weights[j]

				// Write to PENDING voltages, not current.
				// This ensures deterministic behavior regardless of neuron array order.
				state.PendingVoltages[target] += weight
			}
		}
	}
}
