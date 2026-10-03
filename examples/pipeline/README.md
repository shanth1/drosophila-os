# WASM To SNN Pipeline

Run from the repository root:

```sh
make pipeline-lab
```

Or run `TestWASMSignalTriggersOutput` directly in your IDE. Go must be available
in PATH. The test builds its own sensor in `t.TempDir()` and removes the temporary
WASM file after completion, just like the standalone WASM examples.

## Two Files To Read

- `testdata/sensor/main.go`: exports `tick`, imports `env.host_emit_signal`, and
  sends intensity `0.5` to neuron A (index 0) on every call.
- `pipeline_test.go`: builds the sensor, creates two neurons, connects the real
  input ABI to the real engine buffer, and manually advances three steps.

```text
host calls WASM tick()
    -> guest calls env.host_emit_signal(0, float32 bits of 0.5)
    -> production RegisterInputABI decodes the intensity
    -> engine.InputBuffer stores it
host calls engine.Tick()
    -> consumes buffered input and advances the SNN
host checks whether output neuron B fired
```

There is no Manager, timer, goroutine, production sensor, embedded asset, or
external graph. WASI and `_initialize` are still needed for the guest's Go
runtime. For terminology, see `../wasm/README.md` and `../snn/README.md`.

## Expected Steps

The graph is A -> B with weight `1`, both thresholds are `1`, and leak is
disabled. Arrays are ordered `[A, B]`.

| Engine tick | Poll WASM sensor first? | Voltage | Pending | Fired | Output event |
| --- | --- | --- | --- | --- | --- |
| 1 | Yes: send `0.5` | `[0.5, 0]` | `[0, 0]` | `[false, false]` | None |
| 2 | Yes: send `0.5` | `[0, 0]` | `[0, 1]` | `[true, false]` | None |
| 3 | No | `[0, 0]` | `[0, 0]` | `[false, true]` | B fired |

A accumulates two inputs, reaches its threshold on tick 2, resets, and sends
weight `1` to B's pending input. B consumes that input on tick 3 and fires.
The host records that firing as one output event. This is not an effector
plugin or a monitoring alert; it is a minimal observable output.

The sensor's `tick()` and the engine's `Tick()` are distinct operations.
Calling the sensor only buffers a signal; it does not advance the network.
Call the engine between the two sensor polls: the input buffer stores the
latest value per neuron, so two polls before a flush overwrite rather than add.

## Debugging

Useful breakpoints, in execution order:

1. `sensorTick.Call(ctx)` in the example: the host enters WASM.
2. `math.Float32frombits(intensityBits)` in `internal/wasm/abi_input.go`: the
   guest has called back into the native host.
3. `e.Tick()` in the example: step into input flushing and SNN integration.
4. `eventTicks = append(...)` in the example: output occurs on engine tick 3.

The only new connection compared with the separate examples is
`wasm.RegisterInputABI(ctx, runtime, e.Buffer)`: the host ABI now delivers
signals to the engine instead of a test callback.
