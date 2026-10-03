# LEARNING: Plasticity, Feedback APIs & Evolution

## 1. Internal OS APIs (Feedback & Explainability)
Drosophila.OS exposes internal APIs (`internal/api`) specifically designed for human-in-the-loop interaction, completely separate from infrastructure monitoring.

*   **Explainability API (`GET /api/explain`):** IT operations require transparency. When a Motor Neuron (e.g., "Trigger PagerDuty") fires, this API traverses the internal Trace Logs backward, generating a human-readable Root-Cause trace: *"Motor 'Alert' triggered primarily by intense spike in Sensor 'DB_Latency' and gradual increase in 'RAM_Leak'."*
*   **Retroactive Feedback API (`POST /api/feedback`):** If the OS triggers a false positive, a human operator flags it. The system receives `{ "event_id": 1234, "type": "punish" }`, pulls the exact historical state snapshot from BadgerDB, calculates the optimal weight adjustments (Time-Travel STDP), and gradually applies them to the live brain to prevent it from repeating the mistake.

## 2. The Core Mechanism: STDP
Learning uses **Spike-Timing-Dependent Plasticity (STDP)**.
*   If a sensory spike fires *just before* a motor spike, the synaptic weight between them is **increased** (Reinforcement).
*   If the sensor fires *after* or randomly, the weight is **decreased** (Depression).

## 3. Offline Evolution (Genetic Algorithms)
To map the massive `male-cns` connectome to specific IT infrastructure rules without manual coding:
1.  **Ground Truth:** Provide a TSDB log and a JSON file marking historical outage windows.
2.  **Population Simulation:** The engine spawns 100 isolated brain instances with slightly mutated weights in memory.
3.  **Fast-Forward Replay:** Historical logs are streamed through all instances simultaneously.
4.  **Fitness & Crossover:** Brains that spiked the correct Motor Neurons during the outage windows score high. Top performers are crossed over, mutating through N generations until optimal weights are achieved.
