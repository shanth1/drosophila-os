# STORAGE: State Management, Memory & Snapshots

Status: roadmap except for the engine's in-memory SoA/CSR arrays. BadgerDB,
trace logs, snapshots, telemetry persistence, and learned-weight saving are
not implemented; empty storage placeholders have been removed. The current
input buffer is a latest-value mailbox, not the proposed trace ring buffer.

## 1. Storage Philosophy: RAM vs Disk
Drosophila.OS targets ticks every ~10ms and keeps hot data (neuron potentials) in RAM. The roadmap proposes offloading historical/persistent data to a write-optimized database (`BadgerDB`); the current host does not persist this history.

## 2. In-Memory Structures (Hot Path)
*   **Connectome Matrices:** Arrays storing current `Voltage`, `Threshold`, `Decay`, and `Weights`. Optimized using Structure of Arrays (SoA) for CPU cache locality.
*   **Trace Logs (Ring Buffers):** A critical component for learning and context. The system maintains an in-memory circular buffer storing the exact state of all receptors and motor spikes for the last `N` minutes. When the buffer is full, the oldest ticks are overwritten. This buffer is instantly read to generate the `context_ptr` for effector modules when an alert fires.

## 3. Persistent Storage (BadgerDB)
`BadgerDB` (pure Go, LSM-Tree) is the proposed store for a **write-heavy workload**. This choice still needs workload measurements; no database is currently integrated.

### 3.1. Snapshots (Event-Driven Freezes)
When a critical motor action fires (e.g., an Alert is sent), the system captures the context:
1.  The Go engine halts the Ring Buffer for a millisecond.
2.  It serializes the entire Trace Log (the last 5 minutes of input/output history) into a binary payload.
3.  It writes this payload to BadgerDB with a unique `EventID`.
*Purpose:* This snapshot allows the user to retroactively punish or reward the system for an action that happened hours ago.

### 3.2. Continuous Telemetry (Internal TSDB)
For offline learning, the framework continuously writes normalized receptor states (`0.0 - 1.0`) into BadgerDB. The system acts as its own Time-Series Database, decoupling the learning process from external APIs (which often do not provide historical state).

### 3.3. Brain Weights Dump
The learned synaptic weights are saved to BadgerDB (or a flat `.bin` file) periodically or on graceful shutdown, allowing the fly to retain its "memory" across restarts.
