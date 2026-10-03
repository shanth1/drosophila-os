# MODULES: WASM Contracts & Multi-channeling

## 1. Plugin Isolation and Granularity (The UNIX Way)
Drosophila.OS relies on WebAssembly (WASM) for all input/output operations. The host application (Go) runs WASM modules via the `wazero` engine.

**Rule of Granularity:** One module does exactly one job.
Do not create monolithic "ServerAnalyzer" plugins. Create `sensor_cpu.wasm`, `sensor_ram.wasm`, and `action_slack.wasm`. They must not know about each other.

**Unified Loading Strategy:** The engine operates a single unified module pool. Standard plugins are embedded directly into the binary (`//go:embed`), while proprietary plugins are loaded dynamically from the disk (`./plugins/*.wasm`). The host executes them side-by-side with no distinction in privileges.

## 2. Multi-channeling (Critical Concept for Sensors)
A sensor module must never mix different contexts into a single metric.
*Example:* Polling a metric API for "Online Users".
*   If the API returns `5000`, the sensor normalizes this and sends it to the **Vision Receptor**.
*   If the API returns `HTTP 500 Timeout`, the sensor MUST NOT send `0.0` to the Vision Receptor (the system will falsely think users dropped to zero).
*   *Correct Behavior:* The sensor zeroes out the Vision channel (`0.0` - blindness) AND fires a `1.0` signal into a completely separate **Pain Receptor (Nociceptor)** channel.

## 3. ABI Contracts (Application Binary Interface)

To ensure the Go host can run plugins written in any language (Go/TinyGo, Rust, C, Zig), strict C-style function signatures are enforced.

### 3.1. Input ABI (Sensors)
Sensors fetch data, normalize it to a strictly enforced `float32` range of `[0.0, 1.0]`, and pass it to the host.
*   `0.0`: Absolute silence / Minimum possible value.
*   `1.0`: Maximum intensity / Absolute pain / Maximum threshold.

**Contract:**
*   **Exported by WASM:** `tick()` -> The Go host calls this function based on the configured polling interval.
*   **Imported from Host (Go):** `host_emit_signal(receptor_id i32, intensity f32)` -> The WASM module calls this to push the normalized float into the brain's buffer.

### 3.2. Output ABI (Effectors/Actions)
Actions remain completely dormant. They do not poll. They are executed by the Go host only when a specific brain region's electrical potential crosses a configured `threshold`.

**Contract:**
*   **Exported by WASM:** `trigger(action_id i32, intensity f32, context_ptr i32)`
*   **The Context Pointer:** To avoid "Black Box" alerts, when the brain fires a trigger, the Go host serializes a `Trace Summary` (a JSON of the most active receptors over the last seconds) and passes it via `context_ptr`. The effector module parses this context to translate the abstract spike into a human-readable action (e.g., a Slack message saying: *"Panic triggered. Leading causes: Pain_DB, Smell_Network"*).
