package wasm

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// Plugin represents an active, isolated WASM sensor or effector.
type Plugin struct {
	Name     string
	mod      api.Module
	tickFunc api.Function
	interval time.Duration
	stopChan chan struct{}
}

// Manager orchestrates all running WASM instances.
type Manager struct {
	ctx      context.Context
	cancel   context.CancelFunc
	runtime  wazero.Runtime
	receiver SignalReceiver
	mu       sync.Mutex
	plugins  []*Plugin
	errors   chan error
	wg       sync.WaitGroup
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
		ctx:      ctx,
		cancel:   cancel,
		runtime:  runtime,
		receiver: receiver,
		errors:   make(chan error, 1),
	}, nil
}

// LoadSensor instantiates a WASM module from raw bytecode and schedules its polling tick.
func (m *Manager) LoadSensor(name string, wasmBytes []byte, interval time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if interval <= 0 {
		return fmt.Errorf("sensor %s polling interval must be positive", name)
	}

	// Compile the module (wazero optimizes it to native machine code in-memory)
	compiled, err := m.runtime.CompileModule(m.ctx, wasmBytes)
	if err != nil {
		return fmt.Errorf("compile module %s: %w", name, err)
	}
	defer compiled.Close(m.ctx)

	if _, ok := compiled.ExportedFunctions()["_initialize"]; !ok {
		return fmt.Errorf("sensor %s must export _initialize; rebuild with -buildmode=c-shared", name)
	}
	config := wazero.NewModuleConfig().WithName(name).WithStartFunctions("_initialize")
	mod, err := m.runtime.InstantiateModule(m.ctx, compiled, config)
	if err != nil {
		return fmt.Errorf("instantiate module %s: %w", name, err)
	}

	tickFunc := mod.ExportedFunction("tick")
	if tickFunc == nil {
		_ = mod.Close(m.ctx)
		return fmt.Errorf("module %s is missing exported 'tick' function", name)
	}
	if len(tickFunc.Definition().ParamTypes()) != 0 || len(tickFunc.Definition().ResultTypes()) != 0 {
		_ = mod.Close(m.ctx)
		return fmt.Errorf("sensor %s must export tick with signature () -> ()", name)
	}

	plugin := &Plugin{
		Name:     name,
		mod:      mod,
		tickFunc: tickFunc,
		interval: interval,
		stopChan: make(chan struct{}),
	}

	m.plugins = append(m.plugins, plugin)

	// Start isolated polling goroutine
	m.wg.Add(1)
	go m.runPluginLoop(plugin)

	log.Printf("[WASM] Sensor loaded: '%s' (polling every %v)", name, interval)
	return nil
}

func (m *Manager) runPluginLoop(p *Plugin) {
	defer m.wg.Done()
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.stopChan:
			return
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			// Execute the guest tick()
			_, err := p.tickFunc.Call(m.ctx)
			if err != nil {
				if m.ctx.Err() != nil {
					return
				}
				select {
				case m.errors <- fmt.Errorf("sensor %s tick: %w", p.Name, err):
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
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cancel()
	for _, p := range m.plugins {
		close(p.stopChan)
	}
	m.wg.Wait()
	m.plugins = nil
	return m.runtime.Close(context.Background())
}
