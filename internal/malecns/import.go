// Package malecns imports NeuPrint synapse CSV files into binary CSR connectomes.
package malecns

// Stats describes the successfully written connectome.
type Stats struct {
	Neurons  uint32
	Synapses uint32
}

// Import parses a CSV and atomically replaces the output with its binary graph.
// It returns errors without logging or terminating the process.
func Import(csvPath, outPath string) (Stats, error) {
	conn, err := readConnectome(csvPath)
	if err != nil {
		return Stats{}, err
	}
	if err := writeConnectome(outPath, conn); err != nil {
		return Stats{}, err
	}
	return Stats{Neurons: conn.NumNeurons, Synapses: conn.NumEdges}, nil
}
