# Drosophila.OS

A spiking neural network (SNN) prototype with an embedded web UI and an isolated
WebAssembly sensor. This is not yet a monitoring service or a chaos-engineering
tool. The default application demonstrates:

```text
fixed WASM signal -> host ABI -> two-neuron SNN -> output event
```

## Frontend Preview

The first frontend milestone uses React, TypeScript, Vite, and direct Three.js.
It provides one original procedural fly, orbit/zoom controls, basic behavior
previews, and a same-origin cross-tab presentation laboratory. No backend
telemetry or monitoring behavior is connected yet.

Frontend development requires Node.js 22.12+ (or a compatible modern release)
and npm. Start the independent development server:

```sh
make ui-dev
```

Open the URL printed by Vite. Open `/lab` in a neighboring tab to select idle,
working, alarmed, coffee, or break and adjust activity. Controls use
BroadcastChannel and affect presentation only. `/brain` describes planned real
telemetry; it does not fabricate neural activity.

Build and serve the embedded UI:

```sh
make run
# http://127.0.0.1:8080
make run ARGS="-listen 127.0.0.1:8090"
```

The resulting binary always serves the UI alongside the selected backend and
runs until interrupted by default. `make ui-run` is an alias for `make run`.
There is no separate UI mode. All assets are embedded; Node and a CDN are not
needed at runtime. `make build` and `make check` now build frontend assets too.
Before invoking Go builds/tests directly on a fresh checkout, run
`make plugins ui-build`. Architectural decisions and remaining milestones are
documented in `.AGENTS/FRONTEND.md`.

## Quick Start

Requirements: Go 1.25.5 or newer, Make, Node.js, and npm for the combined build. No CGO, TinyGo, dataset, or credentials
are needed for the default test runtime. Run commands from the repository root:

```sh
make run
```

The application uses a fixed sensor sending `0.5` to neuron A, connected to
output neuron B with weight `1`. Both thresholds are `1`; leak is disabled.
The sensor is sampled on every tick, 250 ms apart. Initial steps are:

| Tick | Sensor polled? | Voltage [A, B] | Pending [A, B] | Fired [A, B] |
| --- | --- | --- | --- | --- |
| 1 | Yes | `[0.5, 0]` | `[0, 0]` | `[false, false]` |
| 2 | Yes | `[0, 0]` | `[0, 1]` | `[true, false]` |
| 3 | Yes | `[0.5, 0]` | `[0, 0]` | `[false, true]` |
| 4 | Yes | `[0, 0]` | `[0, 1]` | `[true, false]` |
| 5 | Yes | `[0.5, 0]` | `[0, 0]` | `[false, true]` |

On tick 3 the host logs `msg="output event" tick=3 neuron=1`. This is a test
event, not an incident detector or WASM effector. By default the backend and UI
continue until Ctrl+C. An explicit tick limit stops the whole application:

```sh
make run ARGS="-ticks 6"
make run ARGS="-ticks 0"
make check
```

Output events occur on ticks 3, 5, 7, and so on.
Sensor calls, engine ticks, and output observation share one sequential loop;
the test runtime does not use the asynchronous sensor Manager. Read
`internal/app/test_runtime.go` and `plugins/src/sensor_fixed/main.go` to follow it.
`examples/pipeline` remains an independent deterministic test of the same
signal path, without the application's timer or embedded sensor. The old
three-step demo is no longer an application mode. The current test runtime still
uses a fixed input; a controllable external HTTP service and real HTTP sensor
are the next integration step. Browser presentation is not connected to the
running engine yet.

## Biological Graph Mode

The previous large-graph experiment is still available explicitly:

```sh
make run ARGS="-mode brain -ticks 300"
```

The current local dataset is `data/male_cns.bin`. Data and generated WASM files
are ignored by Git; a fresh checkout does not contain the dataset. If the CSV
already exists, regenerate the graph with:

```sh
go run ./cmd/malecns-importer -csv data/manc_synapses.csv -out data/male_cns.bin
```

The CSV can be obtained separately with `uv run download_brain.py` after setting
`NEUPRINT_TOKEN` in `.env`. This requires network access and NeuPrint credentials;
it is not necessary when the binary graph is already available.

The brain-mode command rebuilds the sensors as WASI reactors, embeds them in `bin/drosophila`, loads
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
weight calibration, not evidence of useful monitoring behavior. Brain mode keeps
these parameters unchanged so startup repair and model tuning remain separate.

Run until Ctrl+C, or choose another graph path:

```sh
make run ARGS="-mode brain -ticks 0"
make run ARGS="-mode brain -brain /absolute/path/male_cns.bin -ticks 100"
make check
```

The random sensor targets neuron index `100`, so its graph needs at least 101
neurons. This index is not a configured biological receptor or incident detector.
Ctrl+C stops HTTP, the host, and its sensor. A sensor failure stops the application with a
nonzero exit status rather than silently reporting success.
Array indices follow the first appearance of biological IDs in the CSV; index
`100` does not identify the same biological neuron across reordered exports.

## Build And Resources

```sh
make build
./bin/drosophila
./bin/drosophila -mode brain -brain /absolute/path/male_cns.bin -ticks 300
./bin/drosophila -help
make clean
```

`make` defaults to `build`. `make build` first compiles the fixed and random
sensors into `plugins/compiled/`, then embeds their bytes in the host binary
at `bin/drosophila`. Generated paths are ignored by Git. `make run` builds
and executes this binary, forwarding `ARGS` as command-line arguments.
`make clean` removes only the host binary and two sensor files, not source, frontend assets, data,
or Go's build cache. WASM stays under `plugins/` because `go:embed` paths cannot
use `../`.

Both WASM sensors are embedded. The default test runtime is self-contained and can run
from any working directory; its graph is constructed in memory. Brain mode
still needs an external graph, whose default path is relative to the working
directory, not the executable.

Always rebuild the sensor after changing its source. `go run ./cmd/drosophila`
alone uses the previously generated WASM. A command-style WASM build without
`-buildmode=c-shared` is incompatible with this loader: it has `_start` instead
of the required `_initialize` and can fail with `runtime.notInitialized`.
On a fresh checkout, run `make plugins ui-build` before building the host or running
`go test ./...` directly: `go:embed` requires generated sensor and frontend files.

### Configuration And Logs

Only CLI flags configure the host: `-mode` defaults to `test` (or select `brain`),
`-brain` defaults to `data/male_cns.bin` and is used only in brain mode,
and `-ticks` defaults to `0` (runs until interrupted). HTTP/UI always uses `-listen`
(default `127.0.0.1:8080`). A positive tick limit stops both HTTP and the backend. For the previous
300-tick large-graph run, specify `-mode brain -ticks 300`. No YAML file or
environment overrides are used. `cmd/drosophila/config.go` parses
flags using a local `FlagSet` and calls `app.Config.Validate`;
`internal/app.Run(ctx, cfg, logger)` validates and executes the application
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

For the default test runtime, read `internal/app/test_runtime.go` first. For brain mode, follow
these files rather than starting with the architecture roadmap:

1. `cmd/drosophila/`: `config.go` parses CLI flags; `main.go` installs signal
   handlers, creates the logger, and calls `app.Run`. `internal/app/run.go`
   supervises HTTP and the selected backend. `internal/app/serve.go` handles HTTP
   serving and bounded graceful shutdown; `internal/app/brain.go` runs the
   biological graph. Runtime configuration is owned by `internal/app/config.go`.
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
8. `cmd/malecns-importer/main.go` handles flags and calls `malecns.Import`.
   `internal/malecns/csv.go` maps biological IDs to dense indices and scales weights
   by `0.05`; `internal/malecns/binary.go` atomically writes the graph.

In brain mode, two clocks are independent: sensor sampling is every 50 ms;
biological ticks are every 10 ms. The test runtime uses one 250 ms clock. These are
scheduling targets, not hard real-time guarantees.

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
embedded fixed/random sensors, WASM input ABI, console activity, and graceful
shutdown. The default test runtime produces deterministic output events from a small
network; the biological graph remains an experimental mode.

Implemented frontend preview: embedded HTTP static serving, procedural fly,
dashboard, and cross-tab visual laboratory. This is not a host telemetry API.

Roadmap only: HTTP host API, WebSocket, test external server, effectors,
proprietary plugin loading from disk, learning, database, and snapshots.
Empty/package-only placeholders for these features and the empty YAML file
have been removed; no implementation was removed. Documents under `.AGENTS/`
distinguish the broader target design from the current runtime. The output ABI
is a draft, with context length and memory ownership still unresolved.

### Code Map And Labs

`cmd/` contains thin CLI entrypoints for the host and offline CSV importer.
`internal/app/` composes the host runtime; `internal/malecns/` owns CSV import and
binary output. `internal/engine/`
contains graph loading and SNN math; `internal/wasm/` contains the input ABI and
sensor lifecycle; `plugins/` contains the embedded sensors and their sources.
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
next-tick propagation, buffer overwrite/reset, embedded WASM lifecycle, and
six steps of the default fixed-sensor pipeline, always-on HTTP in both backend
modes, shared shutdown, component failures, and occupied listen addresses.
They do not establish biological validity, incident-detection quality, or
production performance of the large graph.

Resource limits and model calibration remain future work: the loader checks
format consistency, not a memory budget, and finite weights can still overflow
when accumulated. The current runtime is not a hardened sandbox for untrusted
plugins; no per-plugin memory budget or execution deadline is configured.
