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

This rebuilds the sensor as a WASI reactor, embeds it in `bin/drosophila`, loads
the graph, and runs 300 biological ticks (nominally three seconds). Compilation
and graph loading add startup time. Logs use the standard `slog` text format on
stderr. Expect messages with these fields (timestamps omitted):

```text
level=INFO msg="brain loaded" neurons=... synapses=...
level=INFO msg="sensor loaded" sensor=sensor_random interval=50ms
...
level=INFO msg="engine state" tick=50 spikes=... total_spikes=... target_voltage=...
...
level=INFO msg="application stopped" reason="tick limit" ticks=300 total_spikes=...
```

Counts depend on the imported CSV; they do not establish a complete biological
connectome. Values vary because the sensor is random. A zero instantaneous spike
count does not imply a broken sensor. Voltage resets to zero when a neuron fires.
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
Array indices follow the first appearance of biological IDs in the CSV; index
`100` does not identify the same biological neuron across reordered exports.

## Build And Resources

```sh
make build
./bin/drosophila -brain /absolute/path/male_cns.bin
./bin/drosophila -help
make clean
```

`make` defaults to `build`. `make build` first compiles the sensor into
`plugins/compiled/sensor_random.wasm`, then embeds those bytes in the host binary
at `bin/drosophila`. Both generated paths are ignored by Git. `make run` builds
and executes this binary, forwarding `ARGS` as command-line arguments.
`make clean` removes only these two generated files, not source, data, or Go's
build cache. WASM stays under `plugins/` because `go:embed` paths cannot use `../`.

The WASM sensor is embedded; the graph is still an external file. Thus the host
is one executable, but not yet a fully self-contained distribution. The default
graph path is relative to the working directory, not the executable.

Always rebuild the sensor after changing its source. `go run ./cmd/drosophila`
alone uses the previously generated WASM. A command-style WASM build without
`-buildmode=c-shared` is incompatible with this loader: it has `_start` instead
of the required `_initialize` and can fail with `runtime.notInitialized`.
On a fresh checkout, run `make plugins` before building the host or running
`go test ./...` directly: `go:embed` requires the generated sensor file.

### Configuration And Logs

Only CLI flags configure the host: `-brain` defaults to `data/male_cns.bin`,
and `-ticks` defaults to `300` (`0` runs until interrupted). No YAML file or
environment overrides are used. `cmd/drosophila/config.go` parses and validates
flags using a local `FlagSet`; `run(ctx, cfg, logger)` executes the application
without parsing flags or installing signal handlers.

Help is written to stdout; diagnostics and neural activity statistics are
written to stderr through `log/slog`. Invalid configuration exits with status
`2`; runtime failures exit with status `1`; a completed run or graceful signal
shutdown exits with status `0`. The CLI logs each returned error once. Engine
and WASM packages return errors rather than logging them. Lab examples retain
`t.Log` output. The offline importer also uses `slog` for its diagnostics.

Example log fields include `sensor`, `interval`, `tick`, `spikes`, `total_spikes`,
and `target_voltage`. These describe network activity, not detected incidents.
There is no log-file rotation, custom logging wrapper, or JSON-output flag.

## Follow One Signal

Read these files in order rather than starting with the architecture roadmap:

1. `cmd/drosophila/`: `config.go` parses CLI flags; `main.go` installs signal
   handlers, creates the logger, and runs the graph, sensor, and biological clock.
2. `plugins/src/sensor_random/main.go`: exports `tick()` and sends a random
   `float32` in `[0, 1)` to neuron `100` every 50 ms.
3. `internal/wasm/abi_input.go`: implements `env.host_emit_signal(i32, i32)`.
   The second integer contains IEEE 754 float bits, not an integer conversion.
   Negative neuron IDs and NaN are dropped; intensity is clamped to `[0, 1]`.
4. `internal/engine/buffer.go`: an atomic latest-value mailbox per neuron.
   This is not an event queue: multiple writes before a flush overwrite each
   other. Each biological tick consumes and clears the buffer.
5. `internal/engine/tick.go` and `habituation.go`: apply leak to existing voltage,
   add pending inputs, test the threshold, reset firing neurons, and schedule
   outgoing weighted signals for the next tick.
6. `internal/engine/snn.go`: validates and loads the graph, then initializes
   voltages to zero, thresholds to `1`, and leak rates to `0.05` per tick.
7. `internal/wasm/manager.go`: initializes the WASI reactor, polls `tick()` in a
   separate goroutine, reports failures, and closes the runtime.
8. `cmd/malecns-importer/`: `main.go` handles flags; `csv.go` maps biological IDs
   to dense indices and scales weights by `0.05`; `binary.go` writes the graph.

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

Roadmap only: HTTP host API, dashboard, WebSocket, mock server, effectors,
proprietary plugin loading from disk, learning, database, and snapshots.
Empty/package-only placeholders for these features and the empty YAML file
have been removed; no implementation was removed. Documents under `.AGENTS/`
distinguish the broader target design from the current runtime. The output ABI
is a draft, with context length and memory ownership still unresolved.

### Code Map And Labs

`cmd/` contains the console host and offline CSV importer; `internal/engine/`
contains graph loading and SNN math; `internal/wasm/` contains the input ABI and
sensor lifecycle; `plugins/` contains the embedded sensor and its source.
`download_brain.py` obtains the CSV separately; the Go importer does not call
NeuPrint. Use explicit `-csv` and `-out` paths as shown above.
The defaults are `data/manc_synapses.csv` and `data/male_cns.bin`. Malformed
CSV is rejected with row information. Both the downloader and importer replace
existing output only after successful writes, using temporary paths on the same
filesystem.

The self-contained examples need Go, but no dataset, credentials, or running
service. WASM examples build their own guests in temporary directories rather
than using the production sensor. Each directory has a walkthrough README:

```sh
make wasm-lab      # Guest function calls and guest-to-host callbacks
make snn-lab       # Tiny in-memory graphs: propagation, accumulation, and leak
make pipeline-lab  # Deterministic WASM input -> SNN -> test-side output observation
```

The pipeline lab observes output in the test; it does not implement an effector
ABI or the HTTP host API.

`make check` rebuilds the sensor, runs tests, and runs `go vet`. Tests cover
graph validation, CSV import and binary roundtrips, leak/threshold/reset behavior,
next-tick propagation, buffer overwrite/reset, and embedded WASM lifecycle.
They do not establish biological validity, incident-detection quality, or
production performance of the large graph.

Resource limits and model calibration remain future work: the loader checks
format consistency, not a memory budget, and finite weights can still overflow
when accumulated. The current runtime is not a hardened sandbox for untrusted
plugins; no per-plugin memory budget or execution deadline is configured.
