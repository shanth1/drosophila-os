# ARCHITECTURE: Core Concept, Engine & Build

Status: mixed current implementation and roadmap. The current runtime is a
console SNN prototype with one embedded random sensor, not a monitoring or
chaos-engineering service. Effectors and the self-contained distribution below
are roadmap goals; see README.md for runnable commands and the code map.

## 1. Core Philosophy: Emergent Behavior over Determinism
Drosophila.OS aims to be a biomimetic monitoring and chaos-engineering framework.
It strictly abandons the traditional deterministic IT monitoring approach (`if CPU > 90% then alert`).

Instead, the architecture relies on **Emergence**:
*   **Sensors (Inputs)** do not analyze data. They only translate real-world metrics into normalized electrical impulses (`0.0` - `1.0`).
*   **The Connectome (SNN Engine)** does not know what "CPU" or "HTTP 500" is. It only processes abstract electrical topology using Spiking Neural Network math.
*   **Effectors (Outputs)** do not query states. They sleep until a "Motor Spike" awakens them.

## 2. Biological Connectome (Male CNS Pipeline)
The demo uses an imported biological graph rather than a randomly generated topology. The downloader targets **`male-cns:v1.0` (FlyEM / Janelia)**; graph counts depend on the exported CSV and do not imply biological completeness or fidelity.
*   **Pipeline:** `download_brain.py` retrieves data from NeuPrint into CSV separately. `cmd/malecns-importer/main.go` is an offline CSV-to-binary CLI; it does not query the NeuPrint API.
*   **Translation:** It translates the massive biological graph into flattened, CPU-cache-optimized binary arrays (`.bin`), adhering to Data-Oriented Design constraints.
*   **Loading:** The Go engine loads these flat `.bin` files directly into memory at startup.

## 3. Build Constraints (Zero Dependency Monolith)
The main OS must compile into a **single, portable binary** with zero external dependencies.
This is a distribution goal, not the current packaging: the random WASM sensor
is embedded, but the graph is external and the UI is not implemented. The host
uses Go modules, including `wazero`; zero dependencies does not mean no Go libraries.
*   **Strictly NO CGO:** Cross-compilation (Linux/Windows/macOS/ARM) must work out of the box.
*   **Embedded Assets:** All WASM plugins, 3D UI assets (Three.js), and `.bin` weights MUST be baked into the binary using Go's `//go:embed`.

## 4. The SNN Engine (Data-Oriented Brain)
The engine abandons OOP (no Interface graphs) in favor of Data-Oriented Design (DoD). Memory consists of flat `Voltages`, `Thresholds`, and `Weights` arrays to maximize CPU L1/L2 cache hits during the high-frequency matrix multiplications.

### 4.1. Temporal Decoupling (The Two Clocks)
1.  **Biological Clock (Engine Tick):** Runs continuously at high frequency (~10ms). Computes math, applies leak (decay), and propagates spikes.
2.  **Polling Clock (Sensor Interval):** The random WASM sensor polls every 50ms and writes to an atomic latest-value mailbox per neuron, not a ring buffer or event queue. Writes before a flush overwrite each other; the Biological Clock consumes and clears each value on the next tick.
