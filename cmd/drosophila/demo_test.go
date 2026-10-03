package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestRunDemoPipeline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	// An absent graph path proves the demo does not load any external data.
	cfg := Config{Mode: "demo", BrainPath: "unused.bin", TickLimit: 6}
	if err := run(ctx, cfg, logger); err != nil {
		t.Fatal(err)
	}

	want := []struct {
		polled  bool
		voltage [2]float32
		pending [2]float32
		fired   [2]bool
	}{
		{polled: true, voltage: [2]float32{0.5, 0}},
		{polled: true, pending: [2]float32{0, 1}, fired: [2]bool{true, false}},
		{polled: false, fired: [2]bool{false, true}},
	}
	states, events, stopped := 0, 0, 0
	decoder := json.NewDecoder(&output)
	for {
		var entry struct {
			Message string     `json:"msg"`
			Tick    int        `json:"tick"`
			Polled  bool       `json:"sensor_polled"`
			Voltage [2]float32 `json:"voltage"`
			Pending [2]float32 `json:"pending"`
			Fired   [2]bool    `json:"fired"`
			Ticks   int        `json:"ticks"`
			Events  int        `json:"events"`
			Reason  string     `json:"reason"`
		}
		err := decoder.Decode(&entry)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		switch entry.Message {
		case "demo state":
			expected := want[states%3]
			states++
			if entry.Tick != states || entry.Polled != expected.polled || entry.Voltage != expected.voltage ||
				entry.Pending != expected.pending || entry.Fired != expected.fired {
				t.Fatalf("unexpected state on tick %d: %+v; want %+v", states, entry, expected)
			}
		case "output event":
			events++
			if entry.Tick != events*3 {
				t.Fatalf("unexpected event tick: %d", entry.Tick)
			}
		case "demo stopped":
			stopped++
			if entry.Ticks != 6 || entry.Events != 2 || entry.Reason != "tick limit" {
				t.Fatalf("unexpected shutdown: %+v", entry)
			}
		}
	}
	if states != 6 || events != 2 || stopped != 1 {
		t.Fatalf("expected 6 states, 2 events, 1 shutdown; got %d, %d, %d", states, events, stopped)
	}
}
