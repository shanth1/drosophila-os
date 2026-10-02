# ENVIRONMENT: 3D Frontend, UI & Mock Infrastructure

## 1. Embedded Mock Server (The Sandbox)
To make Drosophila.OS a self-contained MVP, it includes an embedded HTTP server (`:8080/api/metrics`) that mimics a typical web infrastructure (Nodes, Load Balancer, DB, Storage).
*   **Stable State:** By default, it returns normal, slightly noisy metrics (e.g., CPU 40%, 0 errors).
*   **Chaos API:** Endpoints like `/api/mock/kill?target=db` allow users to forcefully inject anomalies into the mock metrics. This is crucial for demonstrating the emergent reactions of the SNN.

## 2. Event-Driven Visualization (WebSockets)
The 3D Frontend (Three.js) is entirely decoupled from decision-making.
*   **No Command Logic:** The Go engine NEVER sends commands like `show_error_animation()`.
*   **Raw State Broadcast:** The Go engine broadcasts a raw JSON representation of the brain's state via WebSockets (e.g., `{"receptor_12_pain": 0.9, "motor_5_panic": 1.0, "global_arousal": 0.8}`).
*   **Frontend Mapping:** The JavaScript frontend parses this JSON and decides how to map it to the 3D scene.

## 3. 3D Scene & UI Mapping Rules
The Three.js scene consists of a cyber-room with server racks and a holographic fly.

*   **Global Lighting (Homeostasis/Arousal):**
    *   Low arousal (Normal) -> Blue/Neon ambient light.
    *   High arousal (Stress) -> Flashing red, alarms.
*   **Local Anomalies (Particles):**
    *   If a specific receptor tied to the "Database" spikes, the 3D model of the Database emits smoke/spark particles.
*   **Avatar Animations (Fly Behavior):**
    *   If the brain is completely dormant (stable infrastructure), the fly triggers a "Smoking/Idle" animation.
    *   If the `Motor_Escape` neuron fires, the fly triggers a frantic flying animation.
*   **Thoughts (Context Overlays):** Minor anomalies trigger floating text bubbles (e.g., *"Node 2 looks heavy"*), providing human-readable context to the SNN's internal state.
