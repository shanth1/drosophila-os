package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"sort"
)

// CSRConnectome represents the Data-Oriented engine-ready memory structure.
type CSRConnectome struct {
	NumNeurons uint32
	NumEdges   uint32
	Offsets    []uint32  // Length: NumNeurons + 1
	Targets    []uint32  // Length: NumEdges
	Weights    []float32 // Length: NumEdges
}

func TranslateToCSR(graph *BioGraph) *CSRConnectome {
	numNeurons := uint32(len(graph.Nodes))
	numEdges := uint32(len(graph.Edges))

	// Sort edges primarily by PreIdx to group them together
	sort.Slice(graph.Edges, func(i, j int) bool {
		return graph.Edges[i].PreIdx < graph.Edges[j].PreIdx
	})

	csr := &CSRConnectome{
		NumNeurons: numNeurons,
		NumEdges:   numEdges,
		Offsets:    make([]uint32, numNeurons+1),
		Targets:    make([]uint32, numEdges),
		Weights:    make([]float32, numEdges),
	}

	edgeIdx := uint32(0)
	currentPre := uint32(0)

	for _, edge := range graph.Edges {
		// If we advanced to a new source neuron, record the offset
		for currentPre < edge.PreIdx {
			currentPre++
			csr.Offsets[currentPre] = edgeIdx
		}

		csr.Targets[edgeIdx] = edge.PostIdx
		csr.Weights[edgeIdx] = edge.Weight
		edgeIdx++
	}

	// Fill remaining offsets for neurons that have no outgoing edges
	for currentPre < numNeurons {
		currentPre++
		csr.Offsets[currentPre] = edgeIdx
	}

	return csr
}

// WriteToFile serializes the CSR structure into a flat binary file
func (csr *CSRConnectome) WriteToFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Use LittleEndian for high performance on x86/ARM platforms
	order := binary.LittleEndian

	// 1. Header (Magic + Version + N + S)
	f.Write([]byte("DROS"))
	binary.Write(f, order, uint32(1)) // Version 1
	binary.Write(f, order, csr.NumNeurons)
	binary.Write(f, order, csr.NumEdges)

	// 2. Data Arrays
	if err := binary.Write(f, order, csr.Offsets); err != nil {
		return fmt.Errorf("write offsets: %w", err)
	}
	if err := binary.Write(f, order, csr.Targets); err != nil {
		return fmt.Errorf("write targets: %w", err)
	}
	if err := binary.Write(f, order, csr.Weights); err != nil {
		return fmt.Errorf("write weights: %w", err)
	}

	return nil
}
