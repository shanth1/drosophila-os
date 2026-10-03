package wasm

import (
	"context"
	"log"
	"math"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// SignalReceiver abstracts the InputBuffer to decouple WASM from the Engine implementation.
type SignalReceiver interface {
	EmitSignal(neuronID uint32, intensity float32)
}

// RegisterInputABI registers the 'host_emit_signal' function in the "env" module namespace.
func RegisterInputABI(ctx context.Context, runtime wazero.Runtime, receiver SignalReceiver) error {
	_, err := runtime.NewHostModuleBuilder("env").
		NewFunctionBuilder().
		WithFunc(func(ctx context.Context, mod api.Module, receptorID int32, intensityBits uint32) {
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
		Export("host_emit_signal").
		Instantiate(ctx)

	if err != nil {
		log.Printf("[WASM-ABI] Failed to export host_emit_signal: %v", err)
		return err
	}

	return nil
}
