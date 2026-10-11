package wasm

import (
	"context"
	"math"

	"github.com/tetratelabs/wazero"
)

// SignalReceiver abstracts the InputBuffer to decouple WASM from the Engine implementation.
type SignalReceiver interface {
	EmitSignal(neuronID uint32, intensity float32)
}

// RegisterInputABI registers the 'host_emit_signal' function in the "env" module namespace.
func RegisterInputABI(ctx context.Context, runtime wazero.Runtime, receiver SignalReceiver) error {
	return RegisterSensorABI(ctx, runtime, receiver, nil)
}

// HTTPProbe returns status and header latency for a target chosen by the host.
// Status zero represents a transport failure. Guests cannot choose URLs.
type HTTPProbe func(context.Context) (statusCode, elapsedMillis uint32)

// RegisterSensorABI registers signal input and, when granted, the HTTP probe capability.
// host_probe_http() -> i64 packs status in the high 32 bits and milliseconds in the low bits.
func RegisterSensorABI(ctx context.Context, runtime wazero.Runtime, receiver SignalReceiver, probe HTTPProbe) error {
	builder := runtime.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithFunc(func(receptorID int32, intensityBits uint32) {
			if receptorID < 0 {
				return
			}
			intensity := math.Float32frombits(intensityBits)
			if math.IsNaN(float64(intensity)) {
				return
			}
			// Clamp intensity to biologically valid bounds [0.0, 1.0]
			if intensity < 0.0 {
				intensity = 0.0
			} else if intensity > 1.0 {
				intensity = 1.0
			}

			receiver.EmitSignal(uint32(receptorID), intensity)
		}).
		Export("host_emit_signal")
	if probe != nil {
		builder.NewFunctionBuilder().WithFunc(func(ctx context.Context) uint64 {
			status, elapsed := probe(ctx)
			return uint64(status)<<32 | uint64(elapsed)
		}).Export("host_probe_http")
	}
	_, err := builder.Instantiate(ctx)
	return err
}
