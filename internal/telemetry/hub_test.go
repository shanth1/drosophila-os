package telemetry

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
	"time"
)

func testHub(t *testing.T) *Hub {
	t.Helper()
	hub, err := New(State{Mode: "test", Status: "starting"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(hub.Close)
	return hub
}

func decode(t *testing.T, data []byte) Message {
	t.Helper()
	var message Message
	if err := json.Unmarshal(data, &message); err != nil {
		t.Fatal(err)
	}
	return message
}

func TestSnapshotHandoffAndEventOrdering(t *testing.T) {
	hub := testHub(t)
	subscriber, err := hub.Subscribe()
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Unsubscribe(subscriber)
	baseline := decode(t, subscriber.Snapshot)
	state := State{Mode: "test", Status: "running", Brain: Brain{OutputEvents: 1}}
	event := &Event{Name: "neural.output.spike", SchemaVersion: 1, Data: OutputSpike{Tick: 1, Neuron: 2}}
	if err := hub.Publish(state, event); err != nil {
		t.Fatal(err)
	}
	update := decode(t, <-subscriber.Updates)
	output := decode(t, <-subscriber.Updates)
	if update.Type != "state" || output.Type != "event" || update.Sequence != baseline.Sequence+1 || output.Sequence != update.Sequence+1 {
		t.Fatalf("invalid ordering: %+v %+v %+v", baseline, update, output)
	}
	// Reconnecting covers the event in state without replaying the old event.
	reconnected, err := hub.Subscribe()
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Unsubscribe(reconnected)
	snapshot := decode(t, reconnected.Snapshot)
	var restored State
	if err := json.Unmarshal(snapshot.Payload, &restored); err != nil {
		t.Fatal(err)
	}
	if snapshot.Sequence != output.Sequence || restored.Brain.OutputEvents != 1 || snapshot.RunID != baseline.RunID {
		t.Fatalf("invalid recovered snapshot: %+v %+v", snapshot, restored)
	}
	if err := hub.Publish(State{Mode: "test", Status: "running", Brain: Brain{Tick: 2}}, nil); err != nil {
		t.Fatal(err)
	}
	next := decode(t, <-reconnected.Updates)
	if next.Sequence != snapshot.Sequence+1 {
		t.Fatal("snapshot/update boundary has a gap")
	}
}

func TestSlowSubscriberDoesNotBlockOthers(t *testing.T) {
	hub := testHub(t)
	slow, err := hub.Subscribe()
	if err != nil {
		t.Fatal(err)
	}
	fast, err := hub.Subscribe()
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Unsubscribe(fast)
	for tick := 0; tick <= subscriberCapacity; tick++ {
		if err := hub.Publish(State{Mode: "test", Brain: Brain{Tick: uint64(tick)}}, nil); err != nil {
			t.Fatal(err)
		}
		<-fast.Updates
	}
	select {
	case <-slow.Done:
	default:
		t.Fatal("slow subscriber was not evicted")
	}
	select {
	case <-fast.Done:
		t.Fatal("healthy subscriber was evicted")
	default:
	}
}

func TestPublishedStateDoesNotAliasRuntime(t *testing.T) {
	hub := testHub(t)
	state := State{Mode: "test", Modules: []Module{{ID: "original", Capabilities: []string{"original.v1"}}}}
	if err := hub.Publish(state, nil); err != nil {
		t.Fatal(err)
	}
	state.Modules[0].ID = "mutated"
	state.Modules[0].Capabilities[0] = "mutated.v1"
	data, err := hub.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	var snapshot State
	if err := json.Unmarshal(decode(t, data).Payload, &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.Modules[0].ID != "original" || snapshot.Modules[0].Capabilities[0] != "original.v1" {
		t.Fatal("snapshot aliases mutable runtime data")
	}
	if err := hub.Publish(state, &Event{Data: math.NaN()}); err == nil {
		t.Fatal("accepted invalid event data")
	}
	next, _ := hub.Snapshot()
	if !reflect.DeepEqual(data, next) {
		t.Fatal("failed publication partially changed the snapshot")
	}
}

func TestWireFixtureMatchesGoContract(t *testing.T) {
	at := time.Date(2026, 10, 4, 12, 0, 0, 0, time.UTC)
	state := State{
		Mode: "test", Status: "running", Brain: Brain{Neurons: 3, Synapses: 2, Tick: 5},
		Modules:      []Module{{ID: "sensor_http", Kind: "sensor", Status: "running", Capabilities: []string{"http.response.v1", "signal.normalized.v1"}}},
		Entities:     []Entity{{ID: "test-http", Kind: "service", Label: "Test HTTP service"}},
		Observations: []Observation{{Source: "sensor_http", Subject: "test-http", Name: "http.response", SchemaVersion: 1, ObservedAt: at, Data: HTTPResponse{StatusCode: 200, ElapsedMS: 2}}},
	}
	payload, err := encodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	message, err := json.Marshal(Message{APIVersion: 1, RunID: "fixture-run", Sequence: 7, Timestamp: at, Type: "snapshot", Payload: payload})
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("../../testdata/telemetry-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var actual, expected any
	if err := json.Unmarshal(message, &actual); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(fixture, &expected); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatal("Go serialization differs from the shared frontend contract fixture")
	}
}
