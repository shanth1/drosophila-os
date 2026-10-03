# Minimal SNN Example

Run from the repository root:

```sh
make snn-lab
```

Or run `TestSpikePropagation` directly in your IDE.

The test uses the real `internal/engine` with a graph constructed in memory:

```text
external input 1 -> A (index 0) --weight 1--> B (index 1)
```

Both thresholds are `1`. Leak is disabled. No WASM, graph files, timers, or
background goroutines are involved. Each `e.Tick()` advances the network by
one step; no wall-clock time needs to pass.

## Expected State

Arrays below are ordered `[A, B]`.

| Step | Voltage | Pending input | Fired |
| --- | --- | --- | --- |
| After buffering input, before ticking | `[0, 0]` | `[0, 0]` | `[false, false]` |
| After tick 1 | `[0, 0]` | `[0, 1]` | `[true, false]` |
| After tick 2 | `[0, 0]` | `[0, 0]` | `[false, true]` |

The external input buffer is separate from `PendingVoltages`. Tick 1 consumes
the external input, fires A, resets A's voltage, and writes the connection
weight to B's pending input. Tick 2 consumes that pending input and fires B.
B has no outgoing connections. `Fired` describes only the most recent tick,
not whether a neuron has ever fired.

## Terminology

- **SNN (Spiking Neural Network):** a network whose neurons communicate through
  discrete firing events, called spikes.
- **LIF (Leaky Integrate-and-Fire):** a model that accumulates input as voltage,
  leaks voltage over time, and fires when the threshold is reached. Leak is
  deliberately zero in this example.
- **Voltage:** accumulated activity, not a measured physical voltage here.
- **Threshold:** voltage required to fire; equality also fires.
- **Synapse / connection:** a directed edge between neurons.
- **Weight:** the amount added to a target's pending input when the source fires.
- **Tick:** one discrete engine step; a timer can schedule it but is not required.
- **Pending input:** activity waiting to be integrated on the next tick.
- **CSR (Compressed Sparse Row):** the array layout used to store connections.

## The Three Connection Arrays

`Offsets = [0, 1, 1]` means A's outgoing edges occupy `[0:1]` and B's occupy
`[1:1]` (empty). `Targets = [1]` names B as the target of the only edge.
`Weights = [1]` gives that edge its weight. The graph is data; the propagation
algorithm remains the production `engine.Tick` implementation.

For debugging, stop before the first `e.Tick()` and step into
`internal/engine/tick.go`. Inspect `Voltages`, `PendingVoltages`, and `Fired`.

Next experiment: change only the weight to `0.5`. Predict the result before
running: B will accumulate voltage `0.5` on tick 2 but will not fire. The current
assertions will fail intentionally because they describe the weight-1 baseline.
