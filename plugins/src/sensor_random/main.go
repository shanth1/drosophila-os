//go:build wasip1 && wasm

package main

import (
	"math"
	"math/rand"
)

// Declare the host function import from module "env"
//
//go:wasmimport env host_emit_signal
func hostEmitSignal(receptorID int32, intensityBits uint32)

//go:wasmexport tick
func tick() {
	// Generate a random stimulus between 0.0 and 1.0
	val := rand.Float32()

	// Choose a target receptor (e.g. neuron index 100)
	targetNeuron := int32(100)

	// Send signal to Go Host
	hostEmitSignal(targetNeuron, math.Float32bits(val))
}

func main() {
	// Reactor initialization runs before the host calls tick().
}
