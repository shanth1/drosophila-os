package main

import (
	"encoding/binary"
	"encoding/csv"
	"flag"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"time"
)

func main() {
	csvFile := flag.String("csv", "manc_synapses.csv", "Path to NeuPrint synapses CSV")
	outFile := flag.String("out", "male_cns.bin", "Path to output binary connectome")
	flag.Parse()

	log.Printf("Starting Male CNS Import Pipeline from %s...", *csvFile)
	start := time.Now()

	f, err := os.Open(*csvFile)
	if err != nil {
		log.Fatalf("Failed to open CSV: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Read()

	// 1. Streaming CSV reading and mapping 64-bit IDs to contiguous 32-bit indices
	idToIdx := make(map[uint64]uint32)
	var edges []BioEdge
	var nextIdx uint32 = 0

	getOrAddIdx := func(bodyId uint64) uint32 {
		if idx, exists := idToIdx[bodyId]; exists {
			return idx
		}
		idToIdx[bodyId] = nextIdx
		nextIdx++
		return nextIdx - 1
	}

	log.Println("Parsing synapses from CSV...")
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("CSV read error: %v", err)
		}

		preId, _ := strconv.ParseUint(record[0], 10, 64)
		postId, _ := strconv.ParseUint(record[1], 10, 64)
		weight, _ := strconv.ParseFloat(record[2], 32)

		edges = append(edges, BioEdge{
			PreIdx:  getOrAddIdx(preId),
			PostIdx: getOrAddIdx(postId),
			Weight:  float32(weight) * 0.05,
		})
	}

	numNeurons := nextIdx
	numEdges := uint32(len(edges))
	log.Printf("Parsed %d unique neurons and %d synapses in %v", numNeurons, numEdges, time.Since(start))

	// 2. We sort the connections by source (PreIdx) to construct the CSR (Compressed Sparse Row) representation.
	log.Println("Sorting and building CSR arrays (Data-Oriented format)...")
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].PreIdx < edges[j].PreIdx
	})

	offsets := make([]uint32, numNeurons+1)
	targets := make([]uint32, numEdges)
	weights := make([]float32, numEdges)

	edgeIdx := uint32(0)
	currentPre := uint32(0)

	for _, edge := range edges {
		for currentPre < edge.PreIdx {
			currentPre++
			offsets[currentPre] = edgeIdx
		}
		targets[edgeIdx] = edge.PostIdx
		weights[edgeIdx] = edge.Weight
		edgeIdx++
	}
	// Populate the biases for neurons with no outgoing connections.
	for currentPre < numNeurons {
		currentPre++
		offsets[currentPre] = edgeIdx
	}

	// 3. Writing to a binary file
	log.Printf("Writing optimized binary to %s...", *outFile)
	out, err := os.Create(*outFile)
	if err != nil {
		log.Fatalf("Failed to create bin file: %v", err)
	}
	defer out.Close()

	order := binary.LittleEndian
	out.Write([]byte("DROS"))
	binary.Write(out, order, uint32(1)) // Version
	binary.Write(out, order, numNeurons)
	binary.Write(out, order, numEdges)
	binary.Write(out, order, offsets)
	binary.Write(out, order, targets)
	binary.Write(out, order, weights)

	stat, _ := os.Stat(*outFile)
	log.Printf("✅ Success! Binary size: %.2f MB. Total time: %v", float64(stat.Size())/1024/1024, time.Since(start))
}
