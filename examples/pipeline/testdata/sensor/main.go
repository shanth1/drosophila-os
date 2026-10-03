//go:build wasip1 && wasm

package main

import "math"

//go:wasmimport env host_emit_signal
func hostEmitSignal(neuron int32, intensityBits uint32)

//go:wasmexport tick
func tick() {
	hostEmitSignal(0, math.Float32bits(0.5))
}

func main() {}
