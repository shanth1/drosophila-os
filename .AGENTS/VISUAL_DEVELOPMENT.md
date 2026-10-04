# Visual Development Session Handoff

## Read this first

The minimal end-to-end development MVP is complete:

`controlled HTTP system -> host probe -> WASM sensor -> SNN -> versioned API
-> live frontend -> fly reaction`.

A new session can focus on appearance, emotions, props, surrounding objects,
animation, and artistic reactions using the existing backend. Read root
`AGENTS.md`, `.AGENTS/FRONTEND.md`, and `.AGENTS/API.md`, then inspect the actual
source files listed below.
No previous conversation, temporary scripts, downloaded browser, or files outside
the repository are required. Generated dependencies/assets are recreated by Make.

This milestone is a working development foundation. It does not implement
production incident detection, rollback, effectors, chaos engineering, or a
calibrated biological model. Surrounding server/plant/room models and advanced
character behavior are the next visual work, not prerequisites for starting it.

## Product decisions that must survive the session boundary

- Exactly one fly is always the central character of Drosophila.OS.
- Different deployments change the world around it: databases, plants, homes,
  servers, and other observed systems. The shared fly can gain new behaviors.
- Coffee, swearing, smoking, breaks, emotional expression, light, sound, and
  screen effects are frontend artistic interpretations, not backend commands.
- Changing emotions or their visual interpretation must not require new API
  fields such as `flyMood`, `drinkCoffee`, or `screenColor`.
- This is a customizable dashboard/template, not a runtime plugin framework.
  Use normal source modules/configuration; do not introduce iframe isolation,
  dynamic plugin loading, or interchangeable central characters.
- Keep React/TypeScript/Vite and direct Three.js with lightweight FSD. React
  owns controls and pages; Three.js owns frame animation. No per-frame React state.
- Keep backend composition in `internal/app`, import in `internal/malecns`, and
  CLI/process concerns in `cmd`. Reviewable small steps still need proper ownership.
- Explain work in Russian; code, comments, and documentation are English.
- The user reviews and commits each stage. Do not commit automatically. Stop at
  requested review boundaries instead of continuing into unrelated backend work.

## Launch from a fresh checkout

Requirements: Go 1.25.5+, Make, Node.js 22.12+ (or compatible newer Node), npm.
The default test mode needs no biological dataset, credentials, or `.env` values.

For graphics only:

```sh
make ui-dev
```

Open the Vite URL printed in the terminal. Open its `/lab` in a second tab on
the exact same origin, including hostname and port. Select a behavior/activity
to enter manual mode. The host may be absent: an offline indicator is expected,
while manual graphics still work. The dashboard at Go port 8080 and a lab at
Vite port 5173 do not share a BroadcastChannel.

For real signals, start the host before the Vite process:

```sh
# Terminal 1: builds everything, then runs indefinitely
make run

# Terminal 2: independent frontend with hot updates
make ui-dev
```

The host serves UI/API at `http://127.0.0.1:8080` and the observed test system at
`http://127.0.0.1:8081`. Vite proxies `/api`, including WebSocket. Keep `changeOrigin:
false` so backend origin validation works. A different host address can be used:

```sh
make run ARGS="-listen 127.0.0.1:8090 -test-listen 127.0.0.1:8091"
DROSOPHILA_API_URL=http://127.0.0.1:8090 make ui-dev
```

Avoid repeatedly rebuilding the Go host for visual edits; Vite handles them.
Make targets run `npm ci`, so do not reinstall dependencies underneath an active
Vite process. After a previous build, `./bin/drosophila` starts the host directly.
Its embedded UI is the last build, not live source; rebuild for release checks.

## Source map and current contracts

| File | Purpose |
| --- | --- |
| `ui/src/entities/fly/state.ts` | Behavior IDs, `FlyState`, cross-tab state validation |
| `ui/src/entities/fly/model.ts` | Original procedural geometry and basic motion |
| `ui/src/widgets/world-scene/WorldScene.tsx` | Camera, floor/grid, lights, render loop, cleanup |
| `ui/src/features/presentation-rules/liveFly.ts` | Live/manual resolution and artistic mappings |
| `ui/src/features/presentation-rules/liveFly.test.ts` | Override, disconnect, gesture expiry checks |
| `ui/src/features/debug-controls/channel.ts` | Presentation-only BroadcastChannel |
| `ui/src/pages/lab/Lab.tsx` | Manual behavior/activity controls |
| `ui/src/pages/dashboard/Dashboard.tsx` | Scene and honest live/last-known status display |
| `ui/src/pages/brain/Brain.tsx` | Real overview counters and module descriptions |
| `ui/src/features/host-connection/useHostConnection.ts` | React subscription wrapper |
| `ui/src/shared/api/` | Wire validation, sequence tracking, reconnect/staleness |
| `ui/src/app/App.tsx` | Page composition and a 200ms presentation-time refresh |
| `ui/src/app/styles.css` | Current dashboard/layout styling |
| `ui/vite.config.js` | HTTP/WebSocket development proxy |
| `ui/embed.go` | Embedded assets and explicit production page routes |

`FlyState` currently contains `mode: live | manual`, a `behavior`, and normalized
`activity`. Behaviors are `idle`, `working`, `alarmed`, `coffee`, and `break`.
`createFly()` returns `{root, update(timeSeconds, state)}`. Three.js supplies
elapsed animation-loop seconds; this is not the host tick or an action-start time.
The floor lies in XZ, Y is up, and the fly faces roughly negative Z.

The fly is procedural and unrigged. Head/legs/wings are local references; no named
public rig, GLB asset, transition controller, action timeline, or priority scheduler
exists yet. Coffee currently shows/moves a cup; break turns the body; these are
previews, not complete drinking/walking sequences. `alarmed` raises the front legs,
scans with the head, agitates the wings, and adds a rotating red spotlight pair,
a pulsing floor ring, and a warm background. Activity scales alarm movement with
a restless baseline at zero and stronger panic at midrange. Lighting progresses
from yellow to red; the rotating beacon and sound engage at 60% activity, using
the shared threshold in `features/presentation-rules/alarm.ts`. The laboratory
provides opt-in filtered two-tone Web Audio siren playback and a volume slider.
Sound plays from the laboratory tab only
after the user enables it and fades out when the alarm ends. Keep that tab open
when previewing the dashboard.

The scene currently contains the fly, a floor, and a grid. No server, plant,
room, smoke/fire particles, cigarette, or speech system is implemented.
Routing uses pathname checks and ordinary links, not React Router. Adding a new
page requires updating both frontend routing and `ui/embed.go`/route checks.

## Adding or refining an emotion/behavior

1. Keep the stable behavior ID and presentation semantics in frontend code. Add
   a new ID to `behaviors` in `state.ts` when needed; it updates the type, basic
   validator, and lab button list. Validate any new state fields explicitly.
2. Implement a distinct pose/motion in the fly model, or extract a focused
   `features/fly-behavior` controller when sequences/transitions justify it.
   Separate reusable rig/geometry from orchestration; do not create a framework.
3. Make it previewable from `/lab`, including any needed expression/prop controls.
   Debug controls must remain presentation-only and same-origin.
4. Add or change automatic mappings in `liveFly.ts` if requested. A visual-only
   behavior may remain manually accessible without a live trigger.
5. Cover meaningful timing, priority, cancellation, or override behavior when
   introduced. Geometry/color edits do not need tests that mirror their literals.

Currently, fresh `neural.output.spike` events trigger a 1.5s alert gesture; repeated
outputs extend the recent-output window. Snapshot counters do not replay gestures.
Other motion uses sampled spikes/neuron count. These are editable artistic choices.
HTTP failure alone is not relabeled as a brain decision. Disconnection is not
healthy operation. Manual mode wins even while new telemetry arrives; "Return to
live telemetry" releases the override.

Do not assume the lab already injects events or can preview arbitrary world effects:
those controls still need implementation. BroadcastChannel state is transient,
without persisted preferences or durable synchronization across page reloads.

## Adding world objects and effects

Use `entities` for geometry/state of independent objects and the world widget
for composition. Presentation rules translate host data into object/character
parameters. Objects must not fetch HTTP, open sockets, or import the fly entity.
Introduce a small build-time factory registry only when multiple kinds need it.

The existing API already describes `test-http` (kind `service`) and `test-network`
(kind `neural-network`). Observation `http.response` v1 has subject `test-http` and
data `{statusCode, elapsedMs, error?}`. An output event has subject `test-network`;
do not silently treat every output as a diagnosed fault on a particular resource.
Represent unknown kinds neutrally or omit a specialized model explicitly.

Bind scene objects by entity ID, not array position or module ID. Resource state
and brain response remain separately visible: smoke based on failed HTTP and a
fly reacting to output are different mappings. Track absence/staleness so a new
or disconnected resource is not displayed as confirmed healthy.

Keep scene instances stable across telemetry updates and update their parameters
instead of recreating geometry. Extend cleanup for any new textures, particle
objects, render targets, timers, or audio resources. Current cleanup handles mesh
and line geometry/materials; it does not automatically cover future textures or
Points effects. Account for React StrictMode mount/unmount and repeated HMR.
Future sound should use a visible mute control and browser user-gesture activation.
Package assets locally; production must not require a CDN.

## Reproduce real reactions

With the host running, use a separate terminal:

```sh
curl http://127.0.0.1:8081/control
curl -X PUT http://127.0.0.1:8081/control -H 'Content-Type: application/json' \
  -d '{"statusCode":503,"delayMs":0}'
curl -X PUT http://127.0.0.1:8081/control -H 'Content-Type: application/json' \
  -d '{"statusCode":200,"delayMs":0}'
curl -X PUT http://127.0.0.1:8081/control -H 'Content-Type: application/json' \
  -d '{"statusCode":200,"delayMs":800}'
```

Use delay 3000ms to exceed the two-second probe timeout. Configuration replaces
the old value; in-flight requests keep their starting snapshot. Neural clocks
are independent, so do not assert exact wall-clock event ticks. After recovery,
already pending propagation and the artistic alert window take time to settle.

## Verification and review workflow

```sh
make ui-build             # Install locked dependencies, type-check, build assets
npm --prefix ui test      # Protocol and presentation semantics
make check               # Frontend build/tests, WASM builds, Go tests, go vet
make build               # Rebuild the self-contained binary
```

For focused UI work, use frontend checks and visual feedback; use full checks
when the change crosses integration boundaries or verifies embedded delivery.
On a fresh checkout, `make plugins ui-build` must precede direct Go tests because
generated files are embedded. `testdata/telemetry-v1.json` is a shared Go/TypeScript
wire fixture, not a mock server. Do not change it merely to add a fly emotion.

The baseline passed `make check` and an actual frontend-client smoke check through
the Go host and Vite proxy, including restart/reconnection. That smoke script was
temporary, not a repository test or required tool. Browser appearance is reviewed
by the user; do not claim visual verification from transport tests. The initial
Chromium download was unavailable, and the user explicitly accepts manual feedback.

Suggested visual verification: orbit/zoom, resize, switch all manual behaviors,
check live/manual precedence under 503 responses, restore 200, test disconnect,
then check the embedded build. Avoid re-running broad checks without a new change
or unresolved failure. Keep changes reviewable, explain them, and leave commits
to the user.
