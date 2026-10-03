package wasm_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func TestWASMHostCallback(t *testing.T) {
	ctx := context.Background()

	// Build a separate guest that imports a function provided by the host.
	wasmPath := filepath.Join(t.TempDir(), "callback.wasm")
	build := exec.CommandContext(ctx, "go", "build", "-mod=readonly",
		"-buildmode=c-shared", "-o", wasmPath, "./testdata/callback")
	build.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm", "CGO_ENABLED=0")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build guest: %v\n%s", err, output)
	}
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatal(err)
	}

	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, runtime); err != nil {
		t.Fatal(err)
	}

	// Make host.report available before loading the guest that imports it.
	var received int32
	calls := 0
	_, err = runtime.NewHostModuleBuilder("host").
		NewFunctionBuilder().
		WithFunc(func(value int32) {
			received = value
			calls++
			t.Logf("Go host received report(%d)", value)
		}).Export("report").Instantiate(ctx)
	if err != nil {
		t.Fatal(err)
	}
	guest, err := runtime.InstantiateWithConfig(ctx, wasmBytes,
		wazero.NewModuleConfig().WithStartFunctions("_initialize"))
	if err != nil {
		t.Fatal(err)
	}

	run := guest.ExportedFunction("run")
	if run == nil {
		t.Fatal("guest is missing exported function run")
	}
	t.Log("Go host calls WASM run()")
	if _, err := run.Call(ctx); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || received != 42 {
		t.Fatalf("expected one report(42), got calls=%d value=%d", calls, received)
	}
	t.Log("WASM run() returned to the Go host")
}
