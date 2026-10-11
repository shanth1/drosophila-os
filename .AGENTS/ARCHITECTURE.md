# ARCHITECTURE: Core Concept, Engine & Build

Status: mixed current implementation and roadmap. The current runtime is a
SNN prototype with always-on HTTP/UI, a controlled external HTTP service and
embedded HTTP sensor in the default three-neuron test runtime, and an embedded random sensor in optional
brain mode, not a monitoring or
chaos-engineering service. Effectors and the fully embedded biological distribution below
are roadmap goals; see README.md for runnable commands and the code map.

## 1. Core Philosophy: Emergent Behavior over Determinism
Drosophila.OS aims to be a biomimetic monitoring and chaos-engineering framework.
It strictly abandons the traditional deterministic IT monitoring approach (`if CPU > 90% then alert`).

Instead, the architecture relies on **Emergence**:
*   **Sensors (Inputs)** do not analyze data. They only translate real-world metrics into normalized electrical impulses (`0.0` - `1.0`).
*   **The Connectome (SNN Engine)** does not know what "CPU" or "HTTP 500" is. It only processes abstract electrical topology using Spiking Neural Network math.
*   **Effectors (Outputs)** do not query states. They sleep until a "Motor Spike" awakens them.

## 2. Biological Connectome (Male CNS Pipeline)
Brain mode uses an imported biological graph; the default test runtime constructs three neurons in memory. The downloader targets **`male-cns:v1.0` (FlyEM / Janelia)**; graph counts depend on the exported CSV and do not imply biological completeness or fidelity.
*   **Pipeline:** `download_brain.py` retrieves data from NeuPrint into CSV separately. `cmd/malecns-importer/main.go` is an offline CSV-to-binary CLI; it does not query the NeuPrint API.
*   **Ownership:** The importer CLI delegates to `internal/malecns.Import`; parsing, CSR construction, encoding, and atomic output replacement live in that package.
*   **Translation:** It translates the massive biological graph into flattened, CPU-cache-optimized binary arrays (`.bin`), adhering to Data-Oriented Design constraints.
*   **Loading:** The Go engine loads these flat `.bin` files directly into memory at startup.

## 3. Build Constraints (Zero Dependency Monolith)
The main OS must compile into a **single, portable binary** with zero external dependencies.
The default test runtime is self-contained: all three WASM sensors are embedded and its graph
is constructed in memory. Brain mode still uses an external graph and the UI
has an embedded procedural fly and cross-tab laboratory, with live host overview
telemetry through HTTP/WebSocket (see `FRONTEND.md` and `API.md`). The host
uses `wazero` and `github.com/coder/websocket` as direct Go dependencies; zero
runtime dependencies does not mean no build-time libraries. Go 1.25.5+, Make,
Node.js 22.12+ (or compatible newer Node), and npm are build requirements.
*   **Strictly NO CGO:** Cross-compilation (Linux/Windows/macOS/ARM) must work out of the box.
*   **Embedded Assets:** Default WASM sensors and built UI assets (including Three.js) use `//go:embed`. The test graph is in memory; embedding biological `.bin` weights is a future distribution goal, not the current brain-mode behavior.
*   **Build Order:** `make build` builds sensors and UI before the host; `make check` also runs frontend tests, Go tests, and `go vet`. Before direct Go builds/tests on a fresh checkout, run `make plugins ui-build`. WASM guests are WASI reactors built with `-buildmode=c-shared`.

## 4. The SNN Engine (Data-Oriented Brain)
Command entrypoints under `cmd/` only handle process/CLI concerns. `internal/app`
owns configuration validation, HTTP/backend composition, and shared runtime
lifetime. Its public `Run(ctx, cfg, logger)` does not parse flags or install signal
handlers. HTTP serving, biological execution, and HTTP-sensor execution are
separate files within this composition package. Engine math and WASM ABI remain
in `internal/engine` and `internal/wasm`. Tests live with their owning package.

`internal/telemetry` stores serialized overview state and ordered live events;
`internal/hostapi` owns HTTP/WebSocket transport. Only the engine owner publishes
snapshots. Slow clients have bounded queues and never block engine ticks. API
handlers do not read mutable neural arrays. The small pure-Go WebSocket dependency
is `github.com/coder/websocket`.

The engine uses Data-Oriented Design (DoD), not object/interface graphs. Neuron
state lives in separate `Voltages`, `PendingVoltages`, `Thresholds`, `DecayRates`, and `Fired` arrays;
outgoing connections use CSR `Offsets`, `Targets`, and `Weights`. Ticks apply
leak/input/threshold checks and propagate firing neurons' weights into pending
inputs for the next tick, not dense matrix multiplication.

### 4.1. Temporal Decoupling (The Two Clocks)
These independent clocks apply to both modes. The default test runtime uses
separate 250ms sensor/engine clocks; HTTP waits do not block neural ticks.
1.  **Biological Clock (Engine Tick):** Runs continuously at high frequency (~10ms). Computes math, applies leak (decay), and propagates spikes.
2.  **Polling Clock (Sensor Interval):** The random WASM sensor polls every 50ms and writes to an atomic latest-value mailbox per neuron, not a ring buffer or event queue. Writes before a flush overwrite each other; the Biological Clock consumes and clears each value on the next tick.
