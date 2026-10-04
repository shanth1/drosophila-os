package wasm

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/shanth1/drosophila-os/plugins"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func TestHTTPSensorNormalization(t *testing.T) {
	ctx := context.Background()
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, runtime); err != nil {
		t.Fatal(err)
	}
	received := make(signalRecorder, 2)
	var status, elapsed uint32
	if err := RegisterSensorABI(ctx, runtime, received, func(context.Context) (uint32, uint32) { return status, elapsed }); err != nil {
		t.Fatal(err)
	}
	guest, err := runtime.InstantiateWithConfig(ctx, plugins.SensorHTTPWASM, wazero.NewModuleConfig().WithStartFunctions("_initialize"))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		status, elapsed  uint32
		latency, failure float32
	}{
		{200, 0, 0, 0}, {204, 20, 0.02, 0}, {200, 800, 0.8, 0}, {200, 5000, 1, 0},
		{302, 1, 0, 1}, {503, 0, 0, 1}, {0, 2000, 0, 1},
	} {
		status, elapsed = test.status, test.elapsed
		if _, err := guest.ExportedFunction("tick").Call(ctx); err != nil {
			t.Fatal(err)
		}
		for neuron, want := range []float32{test.latency, test.failure} {
			select {
			case signal := <-received:
				if signal.neuron != uint32(neuron) || signal.value != want {
					t.Fatalf("%+v: got %+v", test, signal)
				}
			default:
				t.Fatal("HTTP sensor did not emit")
			}
		}
	}
}

func TestHTTPSensorRequiresGrantedCapability(t *testing.T) {
	manager, err := NewManager(context.Background(), make(signalRecorder, 1))
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()
	err = manager.LoadSensor("http", plugins.SensorHTTPWASM, time.Second)
	if err == nil || !strings.Contains(err.Error(), "host_probe_http") {
		t.Fatalf("expected missing probe capability, got %v", err)
	}
}
