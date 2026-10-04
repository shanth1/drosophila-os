# Frontend Architecture and Roadmap

## Current implementation

The initial shell, procedural fly, orbit controls, three routes, basic cross-tab
behavior previews, and embedded UI server are implemented. Run `make ui-dev` for
Vite or `make run` for the binary (`make ui-run` is an alias). HTTP/UI now starts
alongside the backend in every mode. The default `test` backend continuously
samples the embedded fixed sensor into a two-neuron SNN; `brain` retains the
experimental graph/random sensor. There is no `demo` or standalone `ui` mode.
`-ticks` defaults to zero (unlimited); explicit finite runs stop both components.
Failure of either component cancels and joins the other. The `/brain` page is an
explicit placeholder, and frontend telemetry is not connected yet.

The current laboratory controls behavior/activity directly. Event injection,
priority arbitration, full coffee/break sequences, props beyond a cup, and live
telemetry adaptation are still pending. The `live` label currently means the
default visual preview, not an active host subscription. BroadcastChannel state
is transient and same-origin; do not assume durable synchronization across page
reloads or independent browser profiles.

## Product boundaries

Drosophila.OS has exactly one central fly character. Different deployments can
monitor databases, plants, apartments, or other systems; these change the objects
around the fly, not its identity. This is a customizable application template,
not a frontend plugin framework. Do not introduce iframe isolation, runtime
plugin loading, or interchangeable central characters.

Coffee, smoking breaks, swearing, gestures, lighting, sirens, and screen effects
are artistic interpretations of backend state. They are not SNN responsibilities.
The backend may eventually detect anomalies, emit alerts, perform rollbacks,
trigger chaos engineering, or report normal operation. Changing their visual
interpretation must require frontend changes only. These backend capabilities
are roadmap goals, not currently implemented behavior.

## Technology and boundaries

Use React, TypeScript, Vite, and direct Three.js. React owns pages and controls;
Three.js owns scene objects and frame updates. Do not send per-frame animation
through React state. Keep the procedural fly and its animation controller usable
without React or a network connection. A later original rigged GLB asset may
replace procedural geometry while retaining the controller boundary.

Use lightweight Feature-Sliced Design (FSD):

- `app`: routing, global styles, composition.
- `pages`: dashboard (`/`), brain inspector (`/brain`), visual laboratory (`/lab`).
- `widgets`: world scene, debug panel, future brain monitor.
- `features`: fly behavior, presentation rules, debug controls.
- `entities`: fly, future world objects and neural network models.
- `shared`: transport, protocol validation, utilities, UI primitives.

Dependencies point downward; sibling entities do not import one another.
Create slices only when needed, not empty architecture scaffolding. Scene
composition coordinates fly and environment. New environment objects can be
registered with ordinary TypeScript factories at build time.

## Presentation flow

`host message -> validation/adapter -> application event -> presentation rules
-> fly controller and world effects`.

The fly exposes explicit controls for activity, attention, gestures, and
behavior sequences. Its behavior controller arbitrates priorities and prevents
conflicting animations. Urgent behavior can interrupt idle sequences. Repeated
events need coalescing/cooldowns rather than one animation per neural spike.
Backend observations and brain reactions remain distinguishable: an unhealthy
resource must not appear healthy merely because the brain has not reacted.
Disconnection is distinct from health; silence does not imply normal operation.

Start with built-in behaviors such as idle, working, alarmed, coffee, and break.
Configuration handles simple mappings, colors, and thresholds. Use TypeScript
for complex sequences rather than inventing a general-purpose JSON language.

## API direction (not an implemented contract)

Keep transport independent of Three.js and artistic concepts. Useful concepts
are entities, observations, events, actions/results, module descriptions, and
capabilities. Modules and observed entities are distinct, with many-to-many
relationships. New modules must not require changes to the generic envelope.

Use a versioned envelope with sequence, source, subject, namespaced event name,
payload schema version, and data. Describe available signals/actions through
capabilities. Specialized frontend adapters explicitly support known semantics;
unknown modules retain a generic representation. Arbitrary JSON cannot provide
automatic semantic compatibility.

Plan HTTP for snapshots/catalogs/commands and WebSocket for streaming. Select a
small WebSocket dependency when actually implementing transport; do not implement
a custom protocol prematurely. Snapshots and updates need coordinated sequence
numbers, reconnect recovery, and runtime input validation. Replaceable state
updates and significant events need different delivery policies. Slow clients
must never block neural ticks. Capture coherent snapshots in the engine owner,
not by reading mutable arrays concurrently from HTTP handlers.

Rendering, neural ticks, and telemetry have separate clocks. Aggregate/sample
telemetry instead of streaming a full connectome every tick. Protocol versions
and capability payload versions evolve independently. Additive fields are safe;
semantic breaking changes need a new contract or explicit adapter.

## Development and testing

The primary integration environment uses a real Go host, real WASM sensor, and
real SNN against a controllable test HTTP service. This service simulates the
observed world (latency, failures, recovery), not the frontend API. Avoid a
parallel mock backend that duplicates the production API.

`/lab` is a graphics tool: directly invoke gestures or inject presentation events.
Use BroadcastChannel to control a dashboard in another same-origin tab, without
backend involvement. Provide explicit live/manual modes so real input cannot
overwrite debug state. Label manual mode visibly. Test scenarios eventually use
seeds and explicit reset/configuration; deterministic engine checks use controlled
steps rather than assuming wall-clock networking is deterministic. Later, replay
recorded real API messages for presentation regression work.

## Brain page

CSR stores connectivity efficiently; it does not supply anatomical positions.
The current graph file has offsets, targets, and weights, not neuron coordinates.
Never present an algorithmic graph layout as anatomical truth. Anatomical positions
can be added as separate data if available later.

Start `/brain` with spike aggregates, raster plots, selected neuron voltages, and
local connectivity. Request detailed telemetry only when needed. Load topology
separately from activity. Any future 3D graph uses sampling/level of detail and
clearly identifies its layout. The dashboard and brain page observe the same host.

## Build and delivery

Vite serves development with hot updates; later proxy `/api` to Go. Production
assets, including Three.js and models, are local and embedded with `go:embed`.
Node is a build/development dependency only; the delivered Go binary needs no
Node process or CDN. Build frontend assets before compiling the embedding package.
Go serves the SPA routes and embedded resources through `net/http`.

HTTP/UI is an always-on part of application lifetime, not a backend mode.
The three-tick demo has been replaced by continuous fixed-sensor test execution.
The next stage replaces this fixed input with a controllable external HTTP system
and real sensor; `brain` is still experimental, not a production environment mode.

## Implementation milestones

1. React/Vite shell, lightweight FSD, three routes.
2. One original procedural fly, orbit camera, lights, controller boundary.
3. Cross-tab laboratory, basic behaviors and visible debug mode.
4. Embedded static serving, reproducible build commands, usage documentation.
5. First vertical slice: environment object, test external HTTP service, real
   sensor/SNN, versioned host data, frontend adaptation and visible reaction.
6. Real brain telemetry and inspection, starting with aggregates and selections.
7. Additional props, environment objects, behavior sequences, effects and sound.

Milestone 5 is being delivered in reviewable stages: (a) unified HTTP/backend
lifetime and continuous test execution (implemented), (b) controlled external
HTTP service and sensor, (c) minimal real host API/stream, (d) one environment
object and frontend reaction. Stop at each requested review boundary so the user
can inspect, run, and commit the changes. Do not commit without explicit permission.

Milestones 1–4 form the first deliverable. The brain route initially explains
missing telemetry rather than fabricating activity. Backend API, test service,
actual monitoring, and advanced character animation belong to later milestones.
