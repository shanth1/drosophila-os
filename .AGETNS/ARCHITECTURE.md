# ARCHITECTURE: Core Concept, Engine & Build

## 1. Core Philosophy: Emergent Behavior over Determinism
Drosophila.OS is a biomimetic monitoring and chaos-engineering framework.
It strictly abandons the traditional deterministic IT monitoring approach (`if CPU > 90% then alert`).

Instead, the architecture relies on **Emergence**:
*   **Sensors (Inputs)** do not analyze data. They only translate real-world metrics into normalized electrical impulses (`0.0` - `1.0`).
*   **The Connectome (SNN Engine)** does not know what "CPU" or "HTTP 500" is. It only processes abstract electrical topology (pain, smell, vision, temperature) using Spiking Neural Network math.
*   **Effectors (Outputs)** do not query states. They sleep until a "Motor Spike" awakens them.

The system reacts to the *environment's physical properties* (e.g., a sudden drop in "light" combined with a spike in "pain"), allowing it to detect cascading failures without explicit hardcoded rules.

## 2. Build Constraints (Zero Dependency Monolith)
The project must compile into a **single, portable binary** with zero external dependencies.
*   **Strictly NO CGO:** The entire codebase must be pure Go. This ensures cross-compilation (Linux/Windows/macOS/ARM) works out of the box.
*   **Embedded Assets:** All standard WASM plugins, 3D frontend assets (Three.js/HTML/CSS), and pre-trained brain weights (`.bin`) MUST be baked into the binary using Go's `//go:embed` directive.

## 3. Configuration (The Connectome DNA)
Since sensors and the abstract brain are completely decoupled, the system relies on a declarative configuration (e.g., `drosophila.yaml`) to act as the organism's DNA. This config maps specific WASM sensor outputs to exact SNN receptor nodes and sets thresholds for effector modules, binding the abstract math to the physical infrastructure.

## 4. The SNN Engine (Brain)
The core uses a mathematically simplified Leaky Integrate-and-Fire (LIF) model, heavily optimized for Data-Oriented Design (structs of arrays) rather than object-oriented node graphs.

### 4.1. Temporal Decoupling (The Two Clocks)
The system operates on two entirely decoupled time loops:
1.  **Biological Clock (Engine Tick):** Runs continuously at high frequency (e.g., every 10ms). It computes matrix multiplication, applies decay (leak), and propagates spikes across the connectome.
2.  **Polling Clock (Sensor Interval):** Runs asynchronously based on config (e.g., every 15s). Sensors fetch data, normalize it, and write the current `intensity` into a lock-free buffer. The Biological Clock reads from this buffer on its next tick.

### 4.2. Homeostasis (Habituation)
The engine implements biological habituation. The SNN must adapt to background noise. If a receptor receives a constant `0.5` signal for an hour, the downstream neurons adjust their thresholds (raise them) so the network stops spiking. The system is designed to react to **Deltas (sudden spikes or drops)**, not absolute sustained values.
*   **Boiled Frog Prevention (Absolute Nociceptors):** To prevent the system from ignoring slow-growing but fatal anomalies (e.g., a gradual memory leak), the configuration allows flagging specific receptors as *Absolute Nociceptors*. For these nodes, the habituation math (leak/decay) is strictly disabled, ensuring they fire persistently at high thresholds without adaptation.

## 5. High-Level Component Topology
1.  **Engine:** SNN matrix calculator and event loop.
2.  **WASM Manager (`wazero`):** Pure Go WebAssembly runtime. Handles loading, executing, and memory isolation for plugins.
3.  **Storage (`BadgerDB`):** Pure Go LSM-Tree database for high-throughput write-heavy tasks (trace logs, snapshots).
4.  **Broadcaster:** Internal WebSocket server pushing raw JSON state events to the embedded 3D frontend.
5.  **Mock Server:** An embedded HTTP server (`:8080/api/metrics`) providing synthetic infrastructure data for demo/training purposes.
