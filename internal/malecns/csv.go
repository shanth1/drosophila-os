package malecns

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"

	"github.com/shanth1/drosophila-os/internal/engine"
)

func readConnectome(path string) (*engine.Connectome, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open CSV: %w", err)
	}
	conn, parseErr := parseConnectome(f)
	if err := errors.Join(parseErr, f.Close()); err != nil {
		return nil, fmt.Errorf("read CSV %q: %w", path, err)
	}
	return conn, nil
}

func parseConnectome(input io.Reader) (*engine.Connectome, error) {
	reader := csv.NewReader(input)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("CSV row 1: read header: %w", err)
	}
	if len(header) != 3 || header[0] != "pre_bodyId" || header[1] != "post_bodyId" || header[2] != "weight" {
		return nil, fmt.Errorf("CSV row 1: expected exact header pre_bodyId,post_bodyId,weight")
	}
	type edge struct {
		pre, post uint32
		weight    float32
	}
	indices := make(map[uint64]uint32)
	var edges []edge
	maxInt := uint64(^uint(0) >> 1)
	for row := 2; ; row++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("CSV row %d: %w", row, err)
		}
		var ids [2]uint64
		for i := range ids {
			ids[i], err = strconv.ParseUint(record[i], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("CSV row %d: %s: %w", row, header[i], err)
			}
		}
		weight, err := strconv.ParseFloat(record[2], 32)
		if err != nil || math.IsNaN(weight) || math.IsInf(weight, 0) {
			return nil, fmt.Errorf("CSV row %d: weight %q must be a finite float32", row, record[2])
		}
		var mapped [2]uint32
		for i, id := range ids {
			idx, exists := indices[id]
			if !exists {
				count := uint64(len(indices)) + 1
				if count > math.MaxUint32 || count+1 > maxInt/4 {
					return nil, fmt.Errorf("CSV row %d: neuron count exceeds uint32 or platform offset allocation limits", row)
				}
				idx = uint32(len(indices))
				indices[id] = idx
			}
			mapped[i] = idx
		}
		count := uint64(len(edges)) + 1
		if count > math.MaxUint32 || count > maxInt/12 {
			return nil, fmt.Errorf("CSV row %d: edge count exceeds uint32 or platform allocation limits", row)
		}
		edges = append(edges, edge{mapped[0], mapped[1], float32(weight) * 0.05})
	}
	if len(edges) == 0 {
		return nil, fmt.Errorf("CSV row 2: no synapses")
	}
	sort.Slice(edges, func(i, j int) bool { return edges[i].pre < edges[j].pre })
	conn := &engine.Connectome{
		NumNeurons: uint32(len(indices)),
		NumEdges:   uint32(len(edges)),
		Offsets:    make([]uint32, len(indices)+1),
		Targets:    make([]uint32, len(edges)),
		Weights:    make([]float32, len(edges)),
	}
	for i, edge := range edges {
		conn.Offsets[int(edge.pre)+1]++
		conn.Targets[i] = edge.post
		conn.Weights[i] = edge.weight
	}
	for i := 1; i < len(conn.Offsets); i++ {
		conn.Offsets[i] += conn.Offsets[i-1]
	}
	return conn, nil
}
