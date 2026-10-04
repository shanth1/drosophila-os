package app

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestRunFixedSensorPipeline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	// An absent graph path proves the test runtime does not load external data.
	cfg := Config{Mode: "test", BrainPath: "unused.bin", TickLimit: 6, ListenAddress: "127.0.0.1:0"}
	if err := Run(ctx, cfg, logger); err != nil {
		t.Fatal(err)
	}

	want := []struct {
		voltage [2]float32
		pending [2]float32
		fired   [2]bool
	}{
		{voltage: [2]float32{0.5, 0}},
		{pending: [2]float32{0, 1}, fired: [2]bool{true, false}},
		{voltage: [2]float32{0.5, 0}, fired: [2]bool{false, true}},
		{pending: [2]float32{0, 1}, fired: [2]bool{true, false}},
		{voltage: [2]float32{0.5, 0}, fired: [2]bool{false, true}},
		{pending: [2]float32{0, 1}, fired: [2]bool{true, false}},
	}
	states, events, stopped := 0, 0, 0
	decoder := json.NewDecoder(&output)
	for {
		var entry struct {
			Message string     `json:"msg"`
			Tick    int        `json:"tick"`
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
		case "test runtime state":
			if states >= len(want) {
				t.Fatal("received more states than expected")
			}
			expected := want[states]
			states++
			if entry.Tick != states || entry.Voltage != expected.voltage ||
				entry.Pending != expected.pending || entry.Fired != expected.fired {
				t.Fatalf("unexpected state on tick %d: %+v; want %+v", states, entry, expected)
			}
		case "output event":
			events++
			if entry.Tick != events*2+1 {
				t.Fatalf("unexpected event tick: %d", entry.Tick)
			}
		case "test runtime stopped":
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
