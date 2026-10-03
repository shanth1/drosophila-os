package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
)

// BioGraph represents the intermediate state before CSR translation
type BioGraph struct {
	Nodes   []uint64          // Maps local continuous ID (index) to original biological bodyId
	IdToIdx map[uint64]uint32 // Maps biological bodyId to local continuous ID (0 to N-1)
	Edges   []BioEdge
}

type BioEdge struct {
	PreIdx  uint32  // Firing neuron (local ID)
	PostIdx uint32  // Receiving neuron (local ID)
	Weight  float32 // Synaptic weight
}

// ParseNeuPrintCSV reads standard NeuPrint CSV exports.
// Nodes CSV format expected: bodyId,...
// Edges CSV format expected: pre_bodyId,post_bodyId,weight,...
func ParseNeuPrintCSV(nodesPath, edgesPath string) (*BioGraph, error) {
	graph := &BioGraph{
		IdToIdx: make(map[uint64]uint32),
	}

	// 1. Parse Nodes
	nFile, err := os.Open(nodesPath)
	if err != nil {
		return nil, fmt.Errorf("open nodes file: %w", err)
	}
	defer nFile.Close()

	nReader := csv.NewReader(nFile)
	nReader.Read() // Skip header
	nRecords, err := nReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read nodes: %w", err)
	}

	for _, row := range nRecords {
		bodyId, err := strconv.ParseUint(row[0], 10, 64)
		if err != nil {
			continue // skip invalid lines
		}

		// Assign continuous ID
		localIdx := uint32(len(graph.Nodes))
		graph.Nodes = append(graph.Nodes, bodyId)
		graph.IdToIdx[bodyId] = localIdx
	}

	// 2. Parse Edges (Synapses)
	eFile, err := os.Open(edgesPath)
	if err != nil {
		return nil, fmt.Errorf("open edges file: %w", err)
	}
	defer eFile.Close()

	eReader := csv.NewReader(eFile)
	eReader.Read() // Skip header
	eRecords, err := eReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read edges: %w", err)
	}

	for _, row := range eRecords {
		preId, _ := strconv.ParseUint(row[0], 10, 64)
		postId, _ := strconv.ParseUint(row[1], 10, 64)
		weight, _ := strconv.ParseFloat(row[2], 32)

		// Only add edge if both nodes exist in our index
		preIdx, okPre := graph.IdToIdx[preId]
		postIdx, okPost := graph.IdToIdx[postId]

		if okPre && okPost {
			graph.Edges = append(graph.Edges, BioEdge{
				PreIdx:  preIdx,
				PostIdx: postIdx,
				Weight:  float32(weight),
			})
		}
	}

	return graph, nil
}
