package main

import (
	"flag"
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	csvPath := flag.String("csv", "data/manc_synapses.csv", "Path to NeuPrint synapses CSV")
	outPath := flag.String("out", "data/male_cns.bin", "Path to output binary connectome")
	flag.Parse()
	conn, err := readConnectome(*csvPath)
	if err == nil {
		err = writeConnectome(*outPath, conn)
	}
	if err != nil {
		logger.Error("import failed", "error", err)
		os.Exit(1)
	}
	logger.Info("connectome written", "neurons", conn.NumNeurons, "synapses", conn.NumEdges, "path", *outPath)
}
