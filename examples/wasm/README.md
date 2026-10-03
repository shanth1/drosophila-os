# Minimal WASM Examples

Run from the repository root:

```sh
make wasm-lab
```

Or run either test directly in your IDE. Go must be available in PATH.

## 1. Host Calls Guest

Only two source files matter:

- `testdata/plugin/main.go`: guest code, compiled to WASM; exports `add`.
- `wasm_test.go`: native Go host; builds, loads, and calls the guest.

The test has three steps:

1. Build the guest into a temporary `plugin.wasm` file.
2. Load it into wazero and call `_initialize` to prepare its Go runtime.
3. Call `add(2, 3)` and assert that the returned value is `5`.

WASI provides system functions required by the guest's Go runtime, even though
`add` itself only performs arithmetic. Wazero's `Call` API represents numeric
arguments and results as `uint64` slots; the guest function uses `int32`.

No engine, sensor, manager, embedding, or timers are involved.
The WASM file is rebuilt on each uncached test run and removed after the test.

## 2. Guest Calls Host

Read these two files independently of the first example:

- `testdata/callback/main.go`: imports `host.report` and exports `run`.
- `callback_test.go`: registers a Go callback, loads the guest, and calls `run`.

```text
Go test calls WASM run()
    -> WASM calls host.report(42)
    -> Go callback receives 42
    -> WASM run() returns to the Go test
```

The callback is synchronous: it runs during `run.Call`, not in a background
goroutine. The test checks that exactly one callback delivered the value `42`.

Run only this example:

```sh
go test -v -count=1 -run '^TestWASMHostCallback$' ./examples/wasm
```

For debugging, place a breakpoint on `received = value` in `callback_test.go`.
This is ordinary native Go code, called from the WASM guest.

Optional experiment: change only the guest directive to
`//go:wasmimport host missing_report`. Run the test and observe the loading error:
the host does not provide that import. Restore `host report` afterward. Imports
must be registered before loading the guest, with matching names and signatures.

This is the same mechanism used by the project's sensors to call
`env.host_emit_signal`, without the sensor-specific arguments and engine.

## Terminology

### WASM: The Program Format

**WebAssembly (WASM)** is a portable binary instruction format and execution
model. A `.wasm` file contains compiled code, not Go source or native ARM64/x86
machine code. Despite the name, it does not require a browser.

WASM defines computation, memory, imports, and exports. It does not, by itself,
provide operating-system functions such as reading files or getting the time.

### Wazero: The Executor

**Wazero** is the Go library that executes WASM inside our native Go process.
It loads a `.wasm` file, connects imports to their implementations, and lets us
call exported functions. It is neither the Go compiler nor WASI.

```text
Go source --Go compiler--> .wasm file --wazero--> execution
```

### WASI: The System Interface

**WebAssembly System Interface (WASI)** is a standardized set of interfaces
through which a WASM program can request system services, such as clocks,
random data, or file operations. Think of it as a system API, not an executor
and not another program format.

In these examples, `GOOS=wasip1` tells the Go compiler to target **WASI Preview 1**.
The compiled guest's Go runtime imports functions from
`wasi_snapshot_preview1`. The host supplies their implementations by calling:

```go
wasi_snapshot_preview1.Instantiate(ctx, runtime)
```

Even a simple Go guest exporting `add` includes a Go runtime that needs these
imports. A WASM module produced by another toolchain might not need WASI at all.

Registering WASI does not automatically grant access to the host's directories,
environment variables, or all OS capabilities. Those require explicit host
configuration and support from the chosen WASI implementation.

**The distinction:** WASM describes the program; wazero executes it; WASI
describes standard system functions that the executor can provide to it.

### Names Used In The Code

| Term | Meaning in these examples |
| --- | --- |
| Host | The native Go test running wazero. |
| Guest | The Go program compiled into `.wasm`. |
| Module | The compiled WASM program. |
| Instance | A loaded copy of a module with its own memory and state. |
| Export | A guest function available to the host: `add` or `run`. |
| Import | An external function needed by the guest: `host.report` or WASI functions. |
| Host function | A native Go function registered so the guest can call it. |
| ABI | The agreed module/function names, parameter/result types, and their meaning. |
| Reactor | A module initialized once, then used through repeated exported calls. |
| `_initialize` | The compiler-generated function preparing the Go reactor before exported calls. |

Import and export directions above are from the guest's perspective.
`host.report` is our application-specific import, **not part of WASI**.

## Temporary Files

Each test builds its guest into a separate `t.TempDir()`. Go testing removes
that directory after the test, including normal failures through `t.Fatal`.
Forced process termination may prevent cleanup. The ordinary Go build cache
is separate and remains on disk. `defer runtime.Close(ctx)` releases the loaded
runtime and guest instances in memory.
