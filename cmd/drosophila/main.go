package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shanth1/drosophila-os/internal/engine"
	"github.com/shanth1/drosophila-os/internal/wasm"
	"github.com/shanth1/drosophila-os/plugins"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Drosophila.OS:", err)
		os.Exit(1)
	}
}

func run() (err error) {
	brainPath := flag.String("brain", "data/male_cns.bin", "path to the binary connectome")
	ticks := flag.Int("ticks", 300, "number of biological ticks; 0 runs until Ctrl+C")
	flag.Parse()
	if *ticks < 0 {
		return fmt.Errorf("ticks must be non-negative")
	}
	fmt.Println("Booting Drosophila.OS Core...")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	eng, err := engine.LoadEngine(*brainPath)
	if err != nil {
		return err
	}
	if eng.Conn.NumNeurons <= 100 {
		return fmt.Errorf("random sensor requires neuron index 100; graph has %d neurons", eng.Conn.NumNeurons)
	}
	fmt.Printf("Brain loaded: %d neurons, %d synapses\n", eng.Conn.NumNeurons, eng.Conn.NumEdges)

	wasmManager, err := wasm.NewManager(ctx, eng.Buffer)
	if err != nil {
		return err
	}
	tickNum := 0
	var totalSpikes uint64
	defer func() {
		if closeErr := wasmManager.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close WASM runtime: %w", closeErr))
		}
		// Polling has stopped; include failures from the last in-flight call.
		select {
		case sensorErr := <-wasmManager.Errors():
			err = errors.Join(err, sensorErr)
		default:
		}
		if err == nil && *ticks > 0 && tickNum >= *ticks {
			fmt.Printf("Demo completed: %d ticks, %d total spikes.\n", tickNum, totalSpikes)
		}
	}()

	// Load the embedded random sensor (polls every 50ms).
	err = wasmManager.LoadSensor("sensor_random", plugins.SensorRandomWASM, 50*time.Millisecond)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	fmt.Printf("Biological clock: 10ms; random sensor: 50ms; target: #100; tick limit: %d (0 = unlimited)\n", *ticks)
	fmt.Println("Press Ctrl+C to stop. Spikes describe network activity, not detected incidents.")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Stopped by signal.")
			return nil
		case err := <-wasmManager.Errors():
			return err
		case <-ticker.C:
		}
		tickNum++
		eng.Tick()

		// Count spikes in this tick
		spikes := 0
		for _, fired := range eng.State.Fired {
			if fired {
				spikes++
			}
		}
		totalSpikes += uint64(spikes)

		if tickNum%50 == 0 {
			// Print every 50 ticks (nominally 500ms).
			fmt.Printf("[Tick %5d] Spikes now: %-6d | Total spikes: %d | Voltage #100: %.3f\n",
				tickNum, spikes, totalSpikes, eng.State.Voltages[100])
		}

		if *ticks > 0 && tickNum >= *ticks {
			break
		}
	}

	return nil
}
