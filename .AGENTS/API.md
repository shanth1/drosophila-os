# Host Telemetry API v1

Status: implemented read-only overview API. No host commands, anomaly diagnoses,
effectors, graph topology, selected voltages, or durable event journal are exposed.
The test environment's `/control` is a different external-system API on another
listener. Do not conflate these APIs.

## Ownership and endpoints

- `internal/telemetry`: wire types, immutable serialized state, sequences,
  bounded subscriptions.
- `internal/hostapi`: HTTP/WebSocket transport and connection shutdown.
- `internal/app`: runtime composition and publication from the engine owner.
- `ui/src/shared/api`: validation, ordering, reconnect and staleness handling.
- `ui/src/features/presentation-rules`: frontend-only artistic interpretation.

`GET /api/v1/snapshot` returns an envelope with `type: snapshot`. It is not cached.
`GET /api/v1/stream` upgrades to a read-only WebSocket. The first message is an
authoritative snapshot captured atomically with subscription registration.
Subsequent messages are `state` (full overview replacement) and `event`.
Clients must not send application data. The server retains same-origin checks,
uses bounded writes and control-frame heartbeats, and explicitly closes/joins
WebSockets because HTTP graceful shutdown does not own hijacked connections.

## Envelope

Every message has `apiVersion: 1`, `runId`, `sequence`, `timestamp`, `type`, and
`payload`. Run IDs are newly generated per application instance. Sequence zero
is the initial state; each state update/event advances the sequence. Sequence
numbers, not timestamps, define ordering. The client validates safe integer
representation rather than silently rounding uint64 values.

State/event publication is atomic relative to snapshot registration. An output
state contains the updated counters before the corresponding event. A fresh
snapshot's sequence covers both. There is no historical event replay: counters
and `lastOutputAt` survive reconnect in state, but old events do not restart
animations. Gaps or a run change without a snapshot force a fresh connection.
Duplicates are ignored; unknown message types still advance sequence tracking.
Unsupported API versions/invalid known schemas produce explicit incompatibility.
Unknown fields, namespaced observations/events, and payload schema versions can
be skipped without breaking the generic state view.

## State payload

- `mode`: test or experimental brain (future modes are strings).
- `status`: starting or running.
- `brain`: neuron/synapse counts, tick, spikes on that sampled tick, cumulative
  spikes, cumulative output events, optional `lastOutputAt` (null before output).
- `modules`: ID, kind, status, capability names.
- `entities`: ID, kind, label; modules and entities remain distinct.
- `observations`: source module, subject entity, namespaced name, schema version,
  measurement completion time, and capability-specific data.

HTTP test observations use `name: http.response`, `schemaVersion: 1`, source
`sensor_http`, subject `test-http`, and data `{statusCode, elapsedMs, error?}`.
Status zero means transport failure. Absence of an observation means no completed
measurement yet, not a successful zero or a healthy system. Measurements arrive
through an immutable latest-value mailbox; only the engine loop combines them
with a completed neural tick and publishes state. Measurements and neural ticks
have independent clocks, so a snapshot does not claim that its last observation
was necessarily consumed on that exact tick.

Test mode publishes overview state every 250ms engine tick. Brain mode samples
overview state every ten 10ms ticks (10Hz); spikes are a sampled-tick count, not a
rate over the whole interval. No engine array is shared with HTTP handlers.

## Output event

`name: neural.output.spike`, `schemaVersion: 1`, source `snn`, subject
`test-network`, data `{tick, neuron}`. Output neuron 2 is the small test network's
mapped output. Brain mode has no mapped semantic output yet and does not emit
this event for arbitrary neural activity. Never relabel it as an anomaly,
rollback, or other effect that has not been implemented.

## Backpressure and recovery

Each subscriber has a queue of 64 ordered messages. Publication never waits for
network writes. A full queue evicts that subscriber; the server closes its
connection with a reconnect instruction rather than silently dropping events.
This is a live overview stream, not a reliable/durable action delivery system.
Critical future actions need separate persistence/acknowledgment semantics.

The frontend bootstraps via HTTP and then replaces that baseline with the
WebSocket snapshot. It retries network errors with 0.5–10s backoff, detects stale
running-state streams after six seconds, and keeps old values explicitly marked
as last known while disconnected. Starting-state subscriptions may remain quiet
while the graph loads. Cleanup aborts HTTP, timers, and sockets. Host restarts
reset stream state and fresh output gestures.

Vite proxies `/api` including WebSocket to `DROSOPHILA_API_URL` (default
`http://127.0.0.1:8080`), preserving the browser Host header for origin validation.

## Presentation and verification

Live fly activity is an artistic mapping of sampled spikes/neuron count. Fresh
output events trigger a 1.5s alert gesture, with no replay on snapshots. Manual
lab behavior overrides graphics while telemetry continues. Disconnection does
not represent healthy operation; live motion returns to neutral and UI labels
the connection state. Raw HTTP response state is displayed independently.

`/brain` currently shows real overview counters and generic module descriptions;
it has no anatomical layout, raster plots, or detailed neuron inspection yet.

`testdata/telemetry-v1.json` is shared by Go serialization checks and TypeScript
validation tests. Update it intentionally alongside compatible contract changes.
Frontend tests use Node's native test runner/type stripping, not a mock backend.
