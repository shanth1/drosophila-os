# Minimal WASM Example

Run from the repository root:

```sh
make wasm-lab
```

Or run `TestWASMFunctionCall` directly in your IDE. Go must be available in PATH.

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

No engine, sensor, manager, embedding, callbacks, or timers are involved.
The WASM file is rebuilt on each uncached test run and removed after the test.
