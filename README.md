# Drosophila.OS

A console prototype of a spiking neural network (SNN) driven by an isolated
WebAssembly sensor. This is not yet a monitoring service or a chaos-engineering
tool. The current runnable slice is:

```text
random WASM sensor -> host ABI -> input buffer -> SNN ticks -> console statistics
```

## Quick Start

Requirements: Go 1.25.5 or newer, Make, and a binary connectome. No CGO or TinyGo
is needed. Run commands from the repository root.

The current local dataset is `data/male_cns.bin`. Data and generated WASM files
are ignored by Git; a fresh checkout does not contain the dataset. If the CSV
already exists, regenerate the graph with:

```sh
go run ./cmd/malecns-importer -csv data/manc_synapses.csv -out data/male_cns.bin
```

The CSV can be obtained separately with `uv run download_brain.py` after setting
`NEUPRINT_TOKEN` in `.env`. This requires network access and NeuPrint credentials;
it is not necessary when the binary graph is already available.

```sh
make run
```

This rebuilds the sensor as a WASI reactor, embeds it in the Go host, loads the
graph, and runs 300 biological ticks (nominally three seconds). Compilation and
graph loading add startup time. Expect:

```text
Brain loaded: 173023 neurons, 6287789 synapses
[WASM] Sensor loaded: 'sensor_random' (polling every 50ms)
...
[Tick    50] Spikes now: ... | Total spikes: ... | Voltage #100: ...
...
Demo completed: 300 ticks, ... total spikes.
```

Values vary because the sensor is random. A zero instantaneous spike count does
not imply a broken sensor. Voltage resets to zero when a neuron fires.
With the current graph and default weights, activity can quickly saturate at
roughly 160,000 firing neurons per tick. This is a limitation of the model and
weight calibration, not evidence of useful monitoring behavior. The demo keeps
these parameters unchanged so startup repair and model tuning remain separate.

Run until Ctrl+C, or choose another graph path:

```sh
make run ARGS="-ticks 0"
make run ARGS="-brain /absolute/path/male_cns.bin -ticks 100"
make check
```

The random sensor targets neuron index `100`, so its graph needs at least 101
neurons. This index is not a configured biological receptor or incident detector.
Ctrl+C stops the host and its sensor. A sensor failure stops the demo with a
nonzero exit status rather than silently reporting success.

## Build And Resources

```sh
make plugins
CGO_ENABLED=0 go build -o /tmp/drosophila ./cmd/drosophila
/tmp/drosophila -brain /absolute/path/male_cns.bin
```

The WASM sensor is embedded; the graph is still an external file. Thus the host
is one executable, but not yet a fully self-contained distribution. The default
graph path is relative to the working directory, not the executable.

Always rebuild the sensor after changing its source. `go run ./cmd/drosophila`
alone uses the previously generated WASM. A command-style WASM build without
`-buildmode=c-shared` is incompatible with this loader: it has `_start` instead
of the required `_initialize` and can fail with `runtime.notInitialized`.

## Follow One Signal

Read these files in order rather than starting with the architecture roadmap:

1. `cmd/drosophila/main.go`: loads the graph, starts the sensor and biological
   clock, prints activity, and handles shutdown.
2. `plugins/src/sensor_random/main.go`: exports `tick()` and sends a random
   `float32` in `[0, 1)` to neuron `100` every 50 ms.
3. `internal/wasm/abi_input.go`: implements `env.host_emit_signal(i32, i32)`.
   The second integer contains IEEE 754 float bits, not an integer conversion.
   Negative neuron IDs and NaN are dropped; intensity is clamped to `[0, 1]`.
4. `internal/engine/buffer.go`: atomically stores the latest signal per neuron.
   This is not an event queue: multiple writes before a flush overwrite each
   other. Each biological tick consumes and clears the buffer.
5. `internal/engine/tick.go` and `habituation.go`: apply leak to existing voltage,
   add pending inputs, test the threshold, reset firing neurons, and schedule
   outgoing weighted signals for the next tick.
6. `internal/engine/snn.go`: validates and loads the graph, then initializes
   voltages to zero, thresholds to `1`, and leak rates to `0.05` per tick.
7. `internal/wasm/manager.go`: initializes the WASI reactor, polls `tick()` in a
   separate goroutine, reports failures, and closes the runtime.
8. `cmd/malecns-importer/main.go`: converts CSV biological IDs into dense array
   indices and scales connection weights by `0.05`.

Two clocks are independent: sensor sampling is every 50 ms; biological ticks
are every 10 ms. These are scheduling targets, not hard real-time guarantees.

### Engine Data Model

A neuron stores voltage, threshold, leak rate, and whether it fired this tick.
For each tick, the simplified leaky integrate-and-fire model is:

```text
voltage = voltage * (1 - leak) + pending_input
if voltage >= threshold:
    fire and reset voltage to zero
    send connection weights into neighbors' pending inputs for the next tick
```

State uses separate arrays (Struct of Arrays), rather than one Go object per
neuron. Connections use Compressed Sparse Row (CSR): outgoing edges of neuron
`i` occupy `Offsets[i]:Offsets[i+1]` in parallel `Targets` and `Weights` arrays.
Pending inputs keep propagation independent of neuron iteration order.

The binary format is little-endian: `DROS`, version `uint32(1)`, neuron count,
edge count, `uint32` offsets (neurons + 1), `uint32` targets (edges), and
`float32` weights (edges). The loader rejects incompatible or malformed graphs.

## Scope And Tests

Implemented: graph import/loading, simplified LIF engine, atomic input buffer,
embedded random sensor, WASM input ABI, console activity, and graceful shutdown.

Not implemented: HTTP API, dashboard, WebSocket, effectors, learning, database,
and snapshots. Their source packages are placeholders. `drosophila.yaml` is
currently unused. Documents under `.AGENTS/` describe a broader target design,
not the current runtime.

`make check` rebuilds the sensor, runs tests, and runs `go vet`. Tests cover
graph validation, leak/threshold/reset behavior, next-tick propagation, buffer
overwrite/reset, and repeated signals from the actual embedded WASM reactor.
They do not establish biological validity, incident-detection quality, or
production performance of the large graph.
