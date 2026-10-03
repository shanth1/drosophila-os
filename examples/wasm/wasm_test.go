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

func TestWASMFunctionCall(t *testing.T) {
	ctx := context.Background()

	// Build the guest for WASM, independently of the native Go test.
	wasmPath := filepath.Join(t.TempDir(), "plugin.wasm")
	build := exec.CommandContext(ctx, "go", "build", "-mod=readonly",
		"-buildmode=c-shared", "-o", wasmPath, "./testdata/plugin")
	build.Env = append(os.Environ(), "GOOS=wasip1", "GOARCH=wasm", "CGO_ENABLED=0")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build guest: %v\n%s", err, output)
	}
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatal(err)
	}

	// Load the guest and initialize its Go runtime using WASI.
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, runtime); err != nil {
		t.Fatal(err)
	}
	guest, err := runtime.InstantiateWithConfig(ctx, wasmBytes,
		wazero.NewModuleConfig().WithStartFunctions("_initialize"))
	if err != nil {
		t.Fatal(err)
	}

	// Call the guest's exported function from the host.
	add := guest.ExportedFunction("add")
	if add == nil {
		t.Fatal("guest is missing exported function add")
	}
	results, err := add.Call(ctx, 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0] != 5 {
		t.Fatalf("add(2, 3): expected [5], got %v", results)
	}
	t.Log("WASM add(2, 3) = 5")
}
