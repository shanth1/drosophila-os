package main

import (
	"flag"
	"log"
)

func main() {
	csvPath := flag.String("csv", "data/manc_synapses.csv", "Path to NeuPrint synapses CSV")
	outPath := flag.String("out", "data/male_cns.bin", "Path to output binary connectome")
	flag.Parse()
	conn, err := readConnectome(*csvPath)
	if err == nil {
		err = writeConnectome(*outPath, conn)
	}
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Wrote %d neurons and %d synapses to %s", conn.NumNeurons, conn.NumEdges, *outPath)
}
