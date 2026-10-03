package main

import (
	"fmt"
	"time"

	"github.com/shanth1/drosophila-os/internal/engine"
)

func main() {
	fmt.Println("🚀 Booting Drosophila.OS Core...")

	startLoad := time.Now()
	eng, err := engine.LoadEngine("male_cns.bin")
	if err != nil {
		panic(err)
	}
	fmt.Printf("🧠 Brain loaded in %v: %d Neurons, %d Synapses\n",
		time.Since(startLoad), eng.Conn.NumNeurons, eng.Conn.NumEdges)
	fmt.Println("--------------------------------------------------")

	stimulusCount := uint32(eng.Conn.NumNeurons / 20)
	if stimulusCount == 0 {
		stimulusCount = 1
	}

	ticks := 1000
	fmt.Printf("⏱️  Running SNN Engine for %d ticks...\n", ticks)

	startTick := time.Now()
	totalSpikes := 0

	for t := 0; t < ticks; t++ {
		if t < 10 {
			for i := uint32(0); i < stimulusCount; i++ {
				eng.Buffer.EmitSignal(i, 1.5)
			}
		}

		eng.Tick()

		tickSpikes := 0
		for i := uint32(0); i < eng.Conn.NumNeurons; i++ {
			if eng.State.Fired[i] {
				tickSpikes++
				totalSpikes++
			}
		}

		if t == 0 || (t+1)%100 == 0 {
			fmt.Printf("[Tick %4d] Spikes in this tick: %d\n", t+1, tickSpikes)
		}
	}

	elapsed := time.Since(startTick)
	fmt.Println("--------------------------------------------------")
	fmt.Printf("✅ Benchmark Complete!\n")
	fmt.Printf("Total Time:      %v\n", elapsed)
	fmt.Printf("Time per Tick:   %v\n", elapsed/time.Duration(ticks))
	fmt.Printf("Total Spikes:    %d\n", totalSpikes)
}
