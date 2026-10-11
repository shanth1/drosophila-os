# ENVIRONMENT: 3D Frontend, UI & Mock Infrastructure

Status: a procedural fly dashboard, manual laboratory, real overview telemetry,
HTTP snapshots, and WebSocket streaming are implemented. Test mode runs a
controlled external HTTP service. See `FRONTEND.md` and `API.md` for the current
architecture. A procedural room, city backdrop, plants, sofa, rug, decorative
lighting/haze, server racks, and control console are implemented. These objects
are decorative; binding them to monitored entities and resource-state effects
remains future work.

## 1. Controlled External System
`internal/testenv` simulates the observed HTTP service, not a duplicate host API.
The test backend composes it into the main binary on a separate listener.
`GET /health` is measured by the real host/WASM sensor; `GET/PUT /control`
inspects/changes response status and delay. Real neural execution stays enabled.
More infrastructure scenarios can extend this environment when needed.

## 2. Event-Driven Visualization (WebSockets)
The 3D Frontend (Three.js) is completely decoupled from the decision-making engine.
*   **No Command Logic:** The Go OS NEVER sends explicit commands (like `play_alarm()`).
*   **Overview Broadcast:** Versioned snapshots and events describe measured systems and neural activity. Only the runtime owner publishes state; mutable engine arrays are not exposed.
*   **Frontend Mapping:** The frontend validates messages and interprets them artistically. A current output spike briefly triggers an alert gesture; HTTP failure alone is not a brain diagnosis.

## 3. Current Scene and Future Data Mapping
`ui/src/widgets/world-scene/WorldScene.tsx` composes independent entity factories,
updates rack/console animations, applies camera occlusion fading, and renders
through bloom/output postprocessing. The city uses a fixed evening appearance;
day/night orchestration and moving aircraft are future extensions. Rack displays
and console graphics are decorative, not actual HTTP or neural measurements.
Sound is opt-in and plays from `/lab`: siren, coffee sips, and keyboard clicks.
Fly-owned smoke and props are implemented; resource smoke/fire and speech are not.

The fly is always the central character. Surrounding objects and artistic rules
are customizable in the same dashboard, without a runtime plugin framework.
The following mappings are illustrative future design, not current scene behavior:
*   **Global Lighting (Arousal):** Low arousal -> Blue/Neon ambient; High arousal -> Flashing red.
*   **Local Anomalies (Particles):** Spiking receptors tied to specific physical nodes trigger particle emitters (smoke/sparks) on those 3D models.
*   **Context Overlays:** Minor spikes trigger floating thoughts (e.g., *"Pain increasing in DB zone"*).
