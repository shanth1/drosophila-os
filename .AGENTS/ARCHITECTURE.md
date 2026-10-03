# ARCHITECTURE: Core Concept, Engine & Build

## 1. Core Philosophy: Emergent Behavior over Determinism
Drosophila.OS is a biomimetic monitoring and chaos-engineering framework.
It strictly abandons the traditional deterministic IT monitoring approach (`if CPU > 90% then alert`).

Instead, the architecture relies on **Emergence**:
*   **Sensors (Inputs)** do not analyze data. They only translate real-world metrics into normalized electrical impulses (`0.0` - `1.0`).
*   **The Connectome (SNN Engine)** does not know what "CPU" or "HTTP 500" is. It only processes abstract electrical topology using Spiking Neural Network math.
*   **Effectors (Outputs)** do not query states. They sleep until a "Motor Spike" awakens them.

## 2. Biological Connectome (Male CNS Pipeline)
The system does not use randomly generated neural graphs. The core topology is derived from the **`male-cns:v1.0` (FlyEM / Janelia)** biological dataset, mapping 166,000 exact neurons and their synapses.
*   **`malecns-importer`:** A standalone CLI tool in `cmd/` that queries the NeuPrint API.
*   **Translation:** It translates the massive biological graph into flattened, CPU-cache-optimized binary arrays (`.bin`), adhering to Data-Oriented Design constraints.
*   **Loading:** The Go engine loads these flat `.bin` files directly into memory at startup.

## 3. Build Constraints (Zero Dependency Monolith)
The main OS must compile into a **single, portable binary** with zero external dependencies.
*   **Strictly NO CGO:** Cross-compilation (Linux/Windows/macOS/ARM) must work out of the box.
*   **Embedded Assets:** All WASM plugins, 3D UI assets (Three.js), and `.bin` weights MUST be baked into the binary using Go's `//go:embed`.

## 4. The SNN Engine (Data-Oriented Brain)
The engine abandons OOP (no Interface graphs) in favor of Data-Oriented Design (DoD). Memory consists of flat `Voltages`, `Thresholds`, and `Weights` arrays to maximize CPU L1/L2 cache hits during the high-frequency matrix multiplications.

### 4.1. Temporal Decoupling (The Two Clocks)
1.  **Biological Clock (Engine Tick):** Runs continuously at high frequency (~10ms). Computes math, applies leak (decay), and propagates spikes.
2.  **Polling Clock (Sensor Interval):** WASM sensors write to a lock-free ring buffer. The Biological Clock reads from this buffer on the next tick, cleanly separating I/O from core math.
