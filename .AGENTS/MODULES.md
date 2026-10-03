# MODULES: WASM Contracts & Multi-channeling

Status: the embedded fixed and random sensors, WASI reactor lifecycle, and input ABI are
implemented. Multi-channel application mappings, proprietary dynamic loading,
and effectors are roadmap items. The HTTP host API is not implemented.

## 1. Plugin Isolation and Granularity (The UNIX Way)
Drosophila.OS uses WebAssembly (WASM) for the fixed demo sensor and the random brain-mode sensor. The host application (Go) runs them via the `wazero` engine; WASM effectors are planned.

**Rule of Granularity:** One module does exactly one job.
Do not create monolithic "ServerAnalyzer" plugins. Create `sensor_cpu.wasm`, `sensor_ram.wasm`, and `action_slack.wasm`. They must not know about each other.

**Loading Strategy:** Both sensors are embedded directly into the binary (`//go:embed`). The demo calls its fixed sensor synchronously; brain mode polls its random sensor through Manager. A unified module pool and proprietary plugins loaded dynamically from disk are roadmap goals, not implemented behavior. Their privilege model remains to be designed.

## 2. Multi-channeling (Critical Concept for Sensors)
A sensor module must never mix different contexts into a single metric.
*Example:* Polling a metric API for "Online Users".
*   If the API returns `5000`, the sensor normalizes this and sends it to the **Vision Receptor**.
*   If the API request fails, the sensor must not represent the failure as a successful measurement of zero users.
*   *Proposed Behavior:* Do not emit a measurement on the Vision channel; emit `1.0` on a separate **Pain Receptor (Nociceptor)** channel. Missing data and a measured zero are distinct states; their application mapping remains to be designed.

## 3. ABI Contracts (Application Binary Interface)

To ensure the Go host can run plugins written in any language (Go/TinyGo, Rust, C, Zig), strict C-style function signatures are enforced.

### 3.1. Input ABI (Sensors)
Sensors fetch data, normalize it to a strictly enforced `float32` range of `[0.0, 1.0]`, and pass it to the host.
*   `0.0`: Absolute silence / Minimum possible value.
*   `1.0`: Maximum intensity / Absolute pain / Maximum threshold.

**Contract:**
*   **Exported by WASM:** `tick()` -> The Go host calls this function based on the configured polling interval.
*   **Imported from Host (Go):** `env.host_emit_signal(receptor_id i32, intensity_bits i32)` -> The second argument contains the IEEE 754 bits of a normalized `float32`, not a numeric integer conversion.
*   **Initialization:** Sensors export `_initialize()` and are built as WASI reactors. For standard Go use `GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared`. The host calls `_initialize` once before polling `tick`.

### 3.2. Output ABI (Effectors/Actions)
Draft only; not implemented. The empty output ABI placeholder has been removed.
Actions remain completely dormant. They do not poll. They are executed by the Go host only when a specific brain region's electrical potential crosses a configured `threshold`.

**Proposed Contract (Incomplete):**
*   **Exported by WASM:** `trigger(action_id i32, intensity f32, context_ptr i32)`
*   **The Context Pointer:** To avoid "Black Box" alerts, when the brain fires a trigger, the Go host serializes a `Trace Summary` (a JSON of the most active receptors over the last seconds) and passes it via `context_ptr`. The effector module parses this context to translate the abstract spike into a human-readable action (e.g., a Slack message saying: *"Panic triggered. Leading causes: Pain_DB, Smell_Network"*).

The draft does not specify context length, guest-memory allocation, ownership,
or lifetime. These must be resolved before implementation; `context_ptr` must
refer to guest linear memory, not a Go host pointer.
