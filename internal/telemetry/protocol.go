// Package telemetry owns versioned host observations and bounded subscriptions.
package telemetry

import (
	"encoding/json"
	"time"
)

const APIVersion = 1

// Message wraps state or domain events. RunID distinguishes process restarts.
type Message struct {
	APIVersion int             `json:"apiVersion"`
	RunID      string          `json:"runId"`
	Sequence   uint64          `json:"sequence"`
	Timestamp  time.Time       `json:"timestamp"`
	Type       string          `json:"type"`
	Payload    json.RawMessage `json:"payload"`
}

type State struct {
	Mode         string        `json:"mode"`
	Status       string        `json:"status"`
	Brain        Brain         `json:"brain"`
	Modules      []Module      `json:"modules"`
	Entities     []Entity      `json:"entities"`
	Observations []Observation `json:"observations"`
}

type Brain struct {
	Neurons      uint32     `json:"neurons"`
	Synapses     uint32     `json:"synapses"`
	Tick         uint64     `json:"tick"`
	Spikes       uint32     `json:"spikes"`
	TotalSpikes  uint64     `json:"totalSpikes"`
	OutputEvents uint64     `json:"outputEvents"`
	LastOutputAt *time.Time `json:"lastOutputAt"`
}

type Module struct {
	ID           string   `json:"id"`
	Kind         string   `json:"kind"`
	Status       string   `json:"status"`
	Capabilities []string `json:"capabilities"`
}

type Entity struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Label string `json:"label"`
}

type Observation struct {
	Source        string    `json:"source"`
	Subject       string    `json:"subject"`
	Name          string    `json:"name"`
	SchemaVersion int       `json:"schemaVersion"`
	ObservedAt    time.Time `json:"observedAt"`
	Data          any       `json:"data"`
}

type HTTPResponse struct {
	StatusCode uint32 `json:"statusCode"`
	ElapsedMS  uint32 `json:"elapsedMs"`
	Error      string `json:"error,omitempty"`
}

type Event struct {
	Source        string `json:"source"`
	Subject       string `json:"subject"`
	Name          string `json:"name"`
	SchemaVersion int    `json:"schemaVersion"`
	Data          any    `json:"data"`
}

type OutputSpike struct {
	Tick   uint64 `json:"tick"`
	Neuron uint32 `json:"neuron"`
}
