package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeTestBrain(t *testing.T, neurons uint32) string {
	t.Helper()
	var data bytes.Buffer
	for _, field := range []any{
		[4]byte{'D', 'R', 'O', 'S'}, uint32(1), neurons, uint32(0), make([]uint32, neurons+1),
	} {
		if err := binary.Write(&data, binary.LittleEndian, field); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(t.TempDir(), "brain.bin")
	if err := os.WriteFile(path, data.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunTickLimit(t *testing.T) {
	cfg := Config{BrainPath: writeTestBrain(t, 101), TickLimit: 1}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var output bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&output, nil))
	if err := run(ctx, cfg, logger); err != nil {
		t.Fatal(err)
	}
	for _, message := range []string{
		`msg="brain loaded" neurons=101 synapses=0`,
		`msg="sensor loaded" sensor=sensor_random interval=50ms`,
		`msg="application stopped" reason="tick limit" ticks=1`,
	} {
		if !strings.Contains(output.String(), message) {
			t.Fatalf("missing %q in logs:\n%s", message, output.String())
		}
	}
}

func TestRunRejectsInvalidBrain(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	missing := filepath.Join(t.TempDir(), "missing.bin")
	err := run(context.Background(), Config{BrainPath: missing, TickLimit: 1}, logger)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected missing brain error, got %v", err)
	}
	err = run(context.Background(), Config{BrainPath: writeTestBrain(t, 1), TickLimit: 1}, logger)
	if err == nil || !strings.Contains(err.Error(), "requires neuron index 100") {
		t.Fatalf("expected incompatible sensor target error, got %v", err)
	}
}
