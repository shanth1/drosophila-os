package wasm

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/shanth1/drosophila-os/plugins"
	"github.com/tetratelabs/wazero"
)

type receivedSignal struct {
	neuron uint32
	value  float32
}

type signalRecorder chan receivedSignal

func (r signalRecorder) EmitSignal(neuron uint32, value float32) {
	select {
	case r <- receivedSignal{neuron, value}:
	default:
	}
}

func TestEmbeddedSensor(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	received := make(signalRecorder, 10)
	m, err := NewManager(ctx, received)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := m.LoadSensor("random", plugins.SensorRandomWASM, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	// Several successful calls detect initialization and repeated-call failures.
	for i := 0; i < 3; i++ {
		select {
		case signal := <-received:
			if signal.neuron != 100 || !(signal.value >= 0 && signal.value <= 1) {
				t.Fatalf("invalid signal: %+v", signal)
			}
		case err := <-m.Errors():
			t.Fatal(err)
		case <-time.After(5 * time.Second):
			t.Fatal("sensor did not emit a signal")
		}
	}
}

func TestLoadSensorRejectsInvalidInterval(t *testing.T) {
	m, err := NewManager(context.Background(), make(signalRecorder, 1))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	for _, interval := range []time.Duration{0, -time.Second} {
		if err := m.LoadSensor("random", plugins.SensorRandomWASM, interval); err == nil {
			t.Fatalf("accepted interval %v", interval)
		}
	}
}

func TestLoadSensorRequiresReactor(t *testing.T) {
	m, err := NewManager(context.Background(), make(signalRecorder, 1))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	// A valid empty WASM module has no reactor initialization export.
	err = m.LoadSensor("empty", []byte{0, 97, 115, 109, 1, 0, 0, 0}, time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "-buildmode=c-shared") {
		t.Fatalf("expected actionable reactor error, got %v", err)
	}
}

func TestSensorFailureReported(t *testing.T) {
	// Minimal reactor: _initialize returns normally, tick executes unreachable.
	module := []byte{
		0, 97, 115, 109, 1, 0, 0, 0,
		1, 4, 1, 96, 0, 0,
		3, 3, 2, 0, 0,
		7, 22, 2, 11, '_', 'i', 'n', 'i', 't', 'i', 'a', 'l', 'i', 'z', 'e', 0, 0,
		4, 't', 'i', 'c', 'k', 0, 1,
		10, 8, 2, 2, 0, 11, 3, 0, 0, 11,
	}
	m, err := NewManager(context.Background(), make(signalRecorder, 1))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if err := m.LoadSensor("broken", module, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-m.Errors():
		if !strings.Contains(err.Error(), "sensor broken tick") {
			t.Fatalf("missing sensor context: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("sensor failure was not reported")
	}
}

func TestInputABI(t *testing.T) {
	ctx := context.Background()
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)
	received := make(signalRecorder, 1)
	if err := RegisterInputABI(ctx, runtime, received); err != nil {
		t.Fatal(err)
	}
	// Guest wrapper forwards two i32 arguments to the host ABI.
	module := []byte{
		0, 97, 115, 109, 1, 0, 0, 0,
		1, 6, 1, 96, 2, 127, 127, 0,
		2, 24, 1, 3, 'e', 'n', 'v', 16, 'h', 'o', 's', 't', '_', 'e', 'm', 'i', 't', '_', 's', 'i', 'g', 'n', 'a', 'l', 0, 0,
		3, 2, 1, 0,
		7, 8, 1, 4, 'e', 'm', 'i', 't', 0, 1,
		10, 10, 1, 8, 0, 32, 0, 32, 1, 16, 0, 11,
	}
	guest, err := runtime.Instantiate(ctx, module)
	if err != nil {
		t.Fatal(err)
	}
	emit := guest.ExportedFunction("emit")
	for _, tc := range []struct {
		name   string
		neuron int32
		input  float32
		want   float32
		drop   bool
	}{
		{"normal", 100, 0.5, 0.5, false},
		{"low", 100, -1, 0, false},
		{"high", 100, 2, 1, false},
		{"infinity", 100, float32(math.Inf(1)), 1, false},
		{"negative neuron", -1, 0.5, 0, true},
		{"nan", 100, float32(math.NaN()), 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := emit.Call(ctx, uint64(uint32(tc.neuron)), uint64(math.Float32bits(tc.input))); err != nil {
				t.Fatal(err)
			}
			select {
			case signal := <-received:
				if tc.drop || signal.neuron != uint32(tc.neuron) || signal.value != tc.want {
					t.Fatalf("unexpected signal: %+v", signal)
				}
			default:
				if !tc.drop {
					t.Fatal("signal was dropped")
				}
			}
		})
	}
}

func TestCloseAndLoadAfterClose(t *testing.T) {
	m, err := NewManager(context.Background(), make(signalRecorder, 1))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = m.Close() })
	if err := m.LoadSensor("random", plugins.SensorRandomWASM, time.Millisecond); err != nil {
		t.Fatal(err)
	}
	// Concurrent shutdown calls must be safe and wait for polling to stop.
	closed := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() { closed <- m.Close() }()
	}
	for i := 0; i < 2; i++ {
		select {
		case err := <-closed:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("shutdown did not complete")
		}
	}
	if err := m.LoadSensor("late", plugins.SensorRandomWASM, time.Millisecond); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled manager, got %v", err)
	}
}

func TestCloseDuringInitialization(t *testing.T) {
	// Reactor: _initialize calls test.wait(); tick is a no-op.
	module := []byte{
		0, 97, 115, 109, 1, 0, 0, 0,
		1, 4, 1, 96, 0, 0,
		2, 13, 1, 4, 't', 'e', 's', 't', 4, 'w', 'a', 'i', 't', 0, 0,
		3, 3, 2, 0, 0,
		7, 22, 2, 11, '_', 'i', 'n', 'i', 't', 'i', 'a', 'l', 'i', 'z', 'e', 0, 1,
		4, 't', 'i', 'c', 'k', 0, 2,
		10, 9, 2, 4, 0, 16, 0, 11, 2, 0, 11,
	}
	m, err := NewManager(context.Background(), make(signalRecorder, 1))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// Ensure cleanup can interrupt initialization even if the test fails.
		m.cancel()
		_ = m.Close()
	})
	started := make(chan struct{})
	_, err = m.runtime.NewHostModuleBuilder("test").NewFunctionBuilder().
		WithFunc(func(ctx context.Context) {
			close(started)
			<-ctx.Done()
		}).Export("wait").Instantiate(m.ctx)
	if err != nil {
		t.Fatal(err)
	}
	loaded := make(chan error, 1)
	go func() { loaded <- m.LoadSensor("waiting", module, time.Millisecond) }()
	select {
	case <-started:
	case err := <-loaded:
		t.Fatalf("initialization did not reach wait: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("initialization did not start")
	}
	closed := make(chan error, 1)
	go func() { closed <- m.Close() }()
	select {
	case err := <-closed:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Close could not cancel initialization while loading held the lock")
	}
	select {
	case err := <-loaded:
		if err == nil {
			t.Fatal("expected interrupted initialization error")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("LoadSensor did not return after cancellation")
	}
}

func TestLoadSensorValidatesExportsBeforeInitialization(t *testing.T) {
	// _initialize traps. An invalid tick must be rejected before that trap.
	module := []byte{
		0, 97, 115, 109, 1, 0, 0, 0,
		1, 8, 2, 96, 0, 0, 96, 1, 127, 0,
		3, 3, 2, 0, 1,
		7, 22, 2, 11, '_', 'i', 'n', 'i', 't', 'i', 'a', 'l', 'i', 'z', 'e', 0, 0,
		4, 't', 'i', 'c', 'k', 0, 1,
		10, 8, 2, 3, 0, 0, 11, 2, 0, 11,
	}
	m, err := NewManager(context.Background(), make(signalRecorder, 1))
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	err = m.LoadSensor("invalid", module, time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "tick with signature () -> ()") {
		t.Fatalf("expected tick signature error before guest execution, got %v", err)
	}
}
