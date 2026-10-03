package engine

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func connectomeBytes(t *testing.T, neurons uint32, offsets, targets []uint32, weights []float32) []byte {
	t.Helper()
	var b bytes.Buffer
	b.WriteString("DROS")
	for _, data := range []any{uint32(1), neurons, uint32(len(targets)), offsets, targets, weights} {
		if err := binary.Write(&b, binary.LittleEndian, data); err != nil {
			t.Fatal(err)
		}
	}
	return b.Bytes()
}

func loadTestEngine(t *testing.T, data []byte) (*Engine, error) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "connectome.bin")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	return LoadEngine(path)
}

func TestLoadEngineValidation(t *testing.T) {
	valid := connectomeBytes(t, 2, []uint32{0, 1, 1}, []uint32{1}, []float32{-0.5})
	cases := map[string][]byte{
		"magic":              append([]byte("NOPE"), valid[4:]...),
		"trailing bytes":     append(append([]byte(nil), valid...), 0),
		"zero neurons":       connectomeBytes(t, 0, []uint32{0}, nil, nil),
		"first offset":       connectomeBytes(t, 2, []uint32{1, 1, 1}, []uint32{1}, []float32{1}),
		"last offset":        connectomeBytes(t, 2, []uint32{0, 0, 0}, []uint32{1}, []float32{1}),
		"descending offsets": connectomeBytes(t, 2, []uint32{0, 2, 1}, []uint32{1}, []float32{1}),
		"target bounds":      connectomeBytes(t, 2, []uint32{0, 1, 1}, []uint32{2}, []float32{1}),
	}
	for name, value := range map[string]float32{"nan": float32(math.NaN()), "positive infinity": float32(math.Inf(1)), "negative infinity": float32(math.Inf(-1))} {
		cases[name] = connectomeBytes(t, 2, []uint32{0, 1, 1}, []uint32{1}, []float32{value})
	}
	for name, offset := range map[string]int{"version": 4, "huge neurons": 8, "huge edges": 12} {
		data := append([]byte(nil), valid...)
		binary.LittleEndian.PutUint32(data[offset:], math.MaxUint32)
		cases[name] = data
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if e, err := loadTestEngine(t, data); err == nil || e != nil {
				t.Fatalf("expected rejection, got engine %v, error %v", e, err)
			}
		})
	}
	for size := 0; size < len(valid); size++ {
		if e, err := loadTestEngine(t, valid[:size]); err == nil || e != nil {
			t.Fatalf("accepted truncated file of size %d", size)
		}
	}
	for _, data := range [][]byte{valid, connectomeBytes(t, 1, []uint32{0, 0}, nil, nil)} {
		e, err := loadTestEngine(t, data)
		if err != nil {
			t.Fatal(err)
		}
		if len(e.State.Voltages) != int(e.Conn.NumNeurons) || e.State.Thresholds[0] != 1 || e.State.DecayRates[0] != 0.05 {
			t.Fatal("incorrect initial state")
		}
	}
	e, err := loadTestEngine(t, valid)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(e.Conn.Offsets, []uint32{0, 1, 1}) || e.Conn.Targets[0] != 1 || e.Conn.Weights[0] != -0.5 {
		t.Fatal("valid graph changed")
	}
}

func TestTickLeakThresholdAndPropagation(t *testing.T) {
	e, err := loadTestEngine(t, connectomeBytes(t, 2, []uint32{0, 1, 1}, []uint32{1}, []float32{1}))
	if err != nil {
		t.Fatal(err)
	}
	e.State.Voltages[0] = 1
	e.State.DecayRates[0] = 0.25
	e.Tick()
	if e.State.Voltages[0] != 0.75 || e.State.Fired[0] {
		t.Fatal("leak must precede threshold evaluation")
	}
	e.Buffer.EmitSignal(0, 0.4375)
	e.Tick()
	if !e.State.Fired[0] || e.State.Fired[1] || e.State.Voltages[0] != 0 || e.State.Voltages[1] != 0 || e.State.PendingVoltages[1] != 1 {
		t.Fatal("threshold equality must spike and defer propagation")
	}
	e.Tick()
	if e.State.Fired[0] || !e.State.Fired[1] || e.State.Voltages[1] != 0 || e.State.PendingVoltages[1] != 0 {
		t.Fatal("propagation must fire on the next tick and reset state")
	}
	e.Tick()
	if e.State.Fired[0] || e.State.Fired[1] {
		t.Fatal("fired flags must reset each tick")
	}
}

func TestInputBufferOverwriteAndReset(t *testing.T) {
	b := NewInputBuffer(2)
	pending := []float32{0.25, 0.5}
	b.EmitSignal(0, 0.75)
	b.EmitSignal(0, 0.5)
	b.EmitSignal(2, 1)
	b.Flush(pending)
	if !reflect.DeepEqual(pending, []float32{0.75, 0.5}) {
		t.Fatalf("latest input must add to pending voltage: %v", pending)
	}
	b.Flush(pending)
	if !reflect.DeepEqual(pending, []float32{0.75, 0.5}) {
		t.Fatalf("flush must reset inputs: %v", pending)
	}
	b.EmitSignal(1, 1)
	b.EmitSignal(1, 0)
	b.Flush(pending)
	if pending[1] != 0.5 {
		t.Fatal("zero input must overwrite the previous signal")
	}
}
