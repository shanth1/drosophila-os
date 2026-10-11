package malecns

import (
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/shanth1/drosophila-os/internal/engine"
)

const csvHeader = "pre_bodyId,post_bodyId,weight\n"

func TestMalformedCSV(t *testing.T) {
	for _, input := range []string{
		"", "pre_bodyId,post_bodyId,weight,extra\n", "post_bodyId,pre_bodyId,weight\n",
		csvHeader, csvHeader + "\n", csvHeader + "1,2\n", csvHeader + "1,2,3,4\n",
		csvHeader + "\"1,2,3\n", csvHeader + "x,2,3\n", csvHeader + "1,-2,3\n",
		csvHeader + "18446744073709551616,2,3\n", csvHeader + "1,2,x\n",
		csvHeader + "1,2,NaN\n", csvHeader + "1,2,+Inf\n", csvHeader + "1,2,-Inf\n",
		csvHeader + "1,2,3.5e38\n", csvHeader + "1,2,1\n2,3,no\n",
	} {
		t.Run(input, func(t *testing.T) {
			_, err := parseConnectome(strings.NewReader(input))
			if err == nil || !strings.Contains(err.Error(), "row ") {
				t.Fatalf("expected error with row information, got %v", err)
			}
		})
	}
}

func TestMappingCSRAndRoundtrip(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "edges.csv")
	input := csvHeader + "90,10,2\n70,10,-3\n90,70,0\n18446744073709551615,0,3.4028234663852886e38\n"
	if err := os.WriteFile(csvPath, []byte(input), 0600); err != nil {
		t.Fatal(err)
	}
	conn, err := readConnectome(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	want := &engine.Connectome{
		NumNeurons: 5, NumEdges: 4,
		Offsets: []uint32{0, 2, 2, 3, 4, 4},
		Targets: []uint32{1, 2, 1, 4},
		Weights: []float32{float32(2) * 0.05, 0, float32(-3) * 0.05, float32(math.MaxFloat32) * 0.05},
	}
	if !reflect.DeepEqual(conn, want) {
		t.Fatalf("got %+v, want %+v", conn, want)
	}
	path := filepath.Join(dir, "connectome.bin")
	if err := os.WriteFile(path, []byte("old output"), 0600); err != nil {
		t.Fatal(err)
	}
	stats, err := Import(csvPath, path)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Neurons != want.NumNeurons || stats.Synapses != want.NumEdges {
		t.Fatalf("unexpected import statistics: %+v", stats)
	}
	loaded, err := engine.LoadEngine(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(loaded.Conn, want) {
		t.Fatalf("roundtrip got %+v, want %+v", loaded.Conn, want)
	}
	assertNoTemps(t, dir)
}

type failingWriter struct{ remaining int }

func TestImportFailurePreservesOutput(t *testing.T) {
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "invalid.csv")
	outPath := filepath.Join(dir, "existing.bin")
	if err := os.WriteFile(csvPath, []byte(csvHeader+"1,2,invalid\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outPath, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{csvPath, filepath.Join(dir, "missing.csv")} {
		stats, err := Import(path, outPath)
		if err == nil || stats != (Stats{}) {
			t.Fatalf("expected failed import with no statistics, got %+v, %v", stats, err)
		}
		data, err := os.ReadFile(outPath)
		if err != nil || string(data) != "original" {
			t.Fatalf("destination changed: %q, %v", data, err)
		}
		assertNoTemps(t, dir)
	}
}

func (w *failingWriter) Write(p []byte) (int, error) {
	if len(p) > w.remaining {
		n := w.remaining
		w.remaining = 0
		return n, io.ErrClosedPipe
	}
	w.remaining -= len(p)
	return len(p), nil
}

func TestEncodeWriteErrors(t *testing.T) {
	conn, err := parseConnectome(strings.NewReader(csvHeader + "1,2,1\n"))
	if err != nil {
		t.Fatal(err)
	}
	// Fail in each header field and each CSR array.
	for _, limit := range []int{0, 4, 8, 12, 16, 28, 32} {
		if err := encodeConnectome(&failingWriter{limit}, conn); !errors.Is(err, io.ErrClosedPipe) {
			t.Fatalf("limit %d: expected write error, got %v", limit, err)
		}
	}
}

func TestFailedReplacementPreservesDestination(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing")
	if err := os.Mkdir(path, 0700); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(path, "original.bin")
	if err := os.WriteFile(marker, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	conn, err := parseConnectome(strings.NewReader(csvHeader + "1,2,1\n"))
	if err != nil {
		t.Fatal(err)
	}
	if err := writeConnectome(path, conn); err == nil {
		t.Fatal("expected replacement error")
	}
	data, err := os.ReadFile(marker)
	if err != nil || string(data) != "original" {
		t.Fatalf("destination changed: %q, %v", data, err)
	}
	assertNoTemps(t, dir)
}

func TestFailedWritePreservesExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing.bin")
	if err := os.WriteFile(path, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(dir, 0700); err != nil {
			t.Error(err)
		}
	})
	err := writeConnectome(path, &engine.Connectome{})
	if err == nil {
		t.Skip("environment permits writes to a read-only directory")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "original" {
		t.Fatalf("destination changed: %q, %v", data, err)
	}
	assertNoTemps(t, dir)
}

func assertNoTemps(t *testing.T, dir string) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(dir, ".malecns-importer-*"))
	if err != nil || len(paths) != 0 {
		t.Fatalf("temporary leftovers: %v, %v", paths, err)
	}
}
