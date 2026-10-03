# ENVIRONMENT: 3D Frontend, UI & Mock Infrastructure

Status: roadmap only. The mock server, HTTP endpoints, WebSocket broadcaster,
and frontend described below are not implemented. Their empty placeholders
have been removed; the current host prints console statistics only.

## 1. The Mock Server (Standalone Sandbox Project)
To facilitate testing, chaos engineering, and synthetic training without breaking production, the roadmap proposes a dedicated **Mock Server** (future `cmd/mock-server`).
*   **Independence:** This is a separate, highly extensible project. It can be compiled as an independent binary to simulate complex network infrastructures (Nodes, Load Balancers, Databases).
*   **Embedded Sandbox:** For a zero-config MVP, the main Drosophila binary can import and run the Mock Server internally as a background goroutine.
*   **Chaos API:** Endpoints like `/api/mock/kill?target=db` allow engineers to inject forced anomalies, observing the SNN's emergent reactions to synthetic stress.

## 2. Event-Driven Visualization (WebSockets)
The 3D Frontend (Three.js) is completely decoupled from the decision-making engine.
*   **No Command Logic:** The Go OS NEVER sends explicit commands (like `play_alarm()`).
*   **Raw State Broadcast:** The internal Broadcaster pushes raw JSON representation of the SNN state (e.g., `{"node_pain": 0.9, "node_panic": 1.0}`) via WebSockets.
*   **Frontend Mapping:** The frontend parses the state and updates the 3D scene accordingly.

## 3. 3D Scene & UI Mapping Rules
*   **Global Lighting (Arousal):** Low arousal -> Blue/Neon ambient; High arousal -> Flashing red.
*   **Local Anomalies (Particles):** Spiking receptors tied to specific physical nodes trigger particle emitters (smoke/sparks) on those 3D models.
*   **Context Overlays:** Minor spikes trigger floating thoughts (e.g., *"Pain increasing in DB zone"*).
