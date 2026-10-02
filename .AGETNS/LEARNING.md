# LEARNING: Plasticity, STDP & Genetic Algorithms

## 1. The Core Mechanism: STDP
Learning in Drosophila.OS is governed by **Spike-Timing-Dependent Plasticity (STDP)**.
*   If a sensory spike (e.g., "Error Rate") fires *just before* a motor spike (e.g., "Panic Action"), the synaptic weight between them is **increased** (Reinforcement).
*   If the sensor fires *after*, or randomly, the weight is **decreased** (Depression).

## 2. Online Learning (Retroactive Feedback)
Because a human operator cannot punish a false positive in milliseconds, the system uses a retroactive pipeline:
1.  **Trigger:** An operator presses a "False Positive" button in Slack/UI containing an `EventID`.
2.  **API Call:** `POST /api/feedback { "event_id": 1234, "type": "punish" }`
3.  **Snapshot Retrieval:** The Go engine loads the Trace Log snapshot for `1234` from BadgerDB.
4.  **Time-Travel STDP:** The engine applies the STDP math to the historical data in the snapshot to calculate the exact `Delta` (how the weights *should* have been adjusted).
5.  **Application:** The calculated `Delta` is applied to the *current* live in-memory connectome.

## 3. Offline Learning (Genetic Algorithm)
To train the system on historical data (e.g., logs from last month) without writing explicit rules:
1.  **Ground Truth:** The user provides a `ground_truth.json` file marking when outages actually occurred (e.g., "Outage from 15:00 to 16:00").
2.  **Population:** The engine spawns 100 isolated "fly brains" in memory with slightly mutated weights.
3.  **Fast-Forward Replay:** The engine streams the historical TSDB logs through all 100 brains as fast as the CPU allows.
4.  **Fitness Function:** Brains that fired the "Panic" motor neuron during the `ground_truth` window gain points. Brains that fired randomly lose points.
5.  **Evolution:** Top performers are crossed over and mutated for N generations until an optimal weight matrix is achieved.

## 4. Chaos Mock-Training (Sandbox)
Before deploying to production, the brain can be trained in a sandbox. The embedded Mock Server generates synthetic anomalies (e.g., simulated DDoS). The Genetic Algorithm forces the brain to find the correlation between the synthetic noise and the desired action.
