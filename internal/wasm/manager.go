package wasm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Manager loads isolated WASM sensors and polls each in its own goroutine.
type Manager struct {
	ctx     context.Context
	cancel  context.CancelFunc
	runtime wazero.Runtime
	errors  chan error
	mu      sync.Mutex // Serializes loading with shutdown, including wg.Add/Wait.
	wg      sync.WaitGroup
}

// NewManager creates a new Wazero-backed WASM plugin manager.
func NewManager(ctx context.Context, receiver SignalReceiver) (*Manager, error) {
	ctx, cancel := context.WithCancel(ctx)
	runtime := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithCloseOnContextDone(true))

	// WASI support for plugins that use standard libraries (e.g. rand, time, fmt)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, runtime); err != nil {
		cancel()
		_ = runtime.Close(context.Background())
		return nil, fmt.Errorf("instantiate WASI: %w", err)
	}

	if err := RegisterInputABI(ctx, runtime, receiver); err != nil {
		cancel()
		_ = runtime.Close(context.Background())
		return nil, fmt.Errorf("register input abi: %w", err)
	}

	return &Manager{
		ctx:     ctx,
		cancel:  cancel,
		runtime: runtime,
		errors:  make(chan error, 1),
	}, nil
}

// LoadSensor instantiates a WASM module from raw bytecode and schedules its polling tick.
func (m *Manager) LoadSensor(name string, wasmBytes []byte, interval time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if interval <= 0 {
		return fmt.Errorf("sensor %s polling interval must be positive", name)
	}
	if err := m.ctx.Err(); err != nil {
		return fmt.Errorf("load sensor %s: %w", name, err)
	}

	// Compile the module (wazero optimizes it to native machine code in-memory)
	compiled, err := m.runtime.CompileModule(m.ctx, wasmBytes)
	if err != nil {
		return fmt.Errorf("compile module %s: %w", name, err)
	}
	defer compiled.Close(m.ctx)

	exports := compiled.ExportedFunctions()
	initialize := exports["_initialize"]
	if initialize == nil {
		return fmt.Errorf("sensor %s must export _initialize; rebuild with -buildmode=c-shared", name)
	}
	if len(initialize.ParamTypes()) != 0 || len(initialize.ResultTypes()) != 0 {
		return fmt.Errorf("sensor %s must export _initialize with signature () -> ()", name)
	}
	tick := exports["tick"]
	if tick == nil {
		return fmt.Errorf("module %s is missing exported 'tick' function", name)
	}
	if len(tick.ParamTypes()) != 0 || len(tick.ResultTypes()) != 0 {
		return fmt.Errorf("sensor %s must export tick with signature () -> ()", name)
	}
	config := wazero.NewModuleConfig().WithName(name).WithStartFunctions("_initialize")
	mod, err := m.runtime.InstantiateModule(m.ctx, compiled, config)
	if err != nil {
		return fmt.Errorf("instantiate module %s: %w", name, err)
	}
	if err := m.ctx.Err(); err != nil {
		_ = mod.Close(context.Background())
		return fmt.Errorf("initialize sensor %s: %w", name, err)
	}

	m.wg.Add(1)
	go m.pollSensor(name, mod.ExportedFunction("tick"), interval)

	return nil
}

func (m *Manager) pollSensor(name string, tick api.Function, interval time.Duration) {
	defer m.wg.Done()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			// Execute the guest tick()
			_, err := tick.Call(m.ctx)
			if err != nil {
				if m.ctx.Err() != nil {
					return
				}
				select {
				case m.errors <- fmt.Errorf("sensor %s tick: %w", name, err):
				default:
				}
				return
			}
		}
	}
}

// Errors reports fatal sensor failures to the host.
func (m *Manager) Errors() <-chan error {
	return m.errors
}

// Close gracefully terminates all plugins and runtime resources.
func (m *Manager) Close() error {
	// Cancel before locking so a running guest _initialize can be interrupted.
	m.cancel()
	m.mu.Lock()
	defer m.mu.Unlock()

	m.wg.Wait()
	return m.runtime.Close(context.Background())
}
