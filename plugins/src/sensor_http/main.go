//go:build wasip1 && wasm

package main

import "math"

//go:wasmimport env host_probe_http
func hostProbeHTTP() uint64

//go:wasmimport env host_emit_signal
func hostEmitSignal(neuron int32, intensityBits uint32)

//go:wasmexport tick
func tick() {
	measurement := hostProbeHTTP()
	status := uint32(measurement >> 32)
	elapsedMillis := uint32(measurement)
	intensity := float32(elapsedMillis) / 1000
	if intensity > 1 {
		intensity = 1
	}
	failure := float32(0)
	if status < 200 || status >= 300 {
		// Clear stale latency input; failure is a separate receptor, not zero latency.
		intensity, failure = 0, 1
	}
	hostEmitSignal(0, math.Float32bits(intensity))
	hostEmitSignal(1, math.Float32bits(failure))
}

func main() {}
