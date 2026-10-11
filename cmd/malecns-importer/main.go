package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/shanth1/drosophila-os/internal/malecns"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	csvPath := flag.String("csv", "data/manc_synapses.csv", "Path to NeuPrint synapses CSV")
	outPath := flag.String("out", "data/male_cns.bin", "Path to output binary connectome")
	flag.Parse()
	stats, err := malecns.Import(*csvPath, *outPath)
	if err != nil {
		logger.Error("import failed", "error", err)
		os.Exit(1)
	}
	logger.Info("connectome written", "neurons", stats.Neurons, "synapses", stats.Synapses, "path", *outPath)
}
