package hostapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/shanth1/drosophila-os/internal/telemetry"
)

func TestSnapshotAndStreamRecovery(t *testing.T) {
	hub, err := telemetry.New(telemetry.State{Mode: "test", Status: "starting"})
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Close()
	handler := New(hub)
	server := httptest.NewServer(handler)
	defer server.Close()
	defer handler.Close()
	response, err := server.Client().Get(server.URL + "/api/v1/snapshot")
	if err != nil {
		t.Fatal(err)
	}
	var initial telemetry.Message
	err = json.NewDecoder(response.Body).Decode(&initial)
	response.Body.Close()
	if err != nil || initial.Type != "snapshot" || response.Header.Get("Cache-Control") != "no-store" {
		t.Fatalf("snapshot: %+v, %v", initial, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/stream"
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	read := func(conn *websocket.Conn) telemetry.Message {
		t.Helper()
		kind, data, err := conn.Read(ctx)
		if err != nil {
			t.Fatal(err)
		}
		var message telemetry.Message
		if err := json.Unmarshal(data, &message); err != nil || kind != websocket.MessageText {
			t.Fatalf("stream message: %v", err)
		}
		return message
	}
	baseline := read(conn)
	if baseline.Type != "snapshot" || baseline.RunID != initial.RunID {
		t.Fatal("stream did not start with the authoritative snapshot")
	}
	if err := hub.Publish(telemetry.State{Mode: "test", Status: "running", Brain: telemetry.Brain{OutputEvents: 1}}, &telemetry.Event{Name: "neural.output.spike", SchemaVersion: 1, Data: telemetry.OutputSpike{Tick: 1, Neuron: 2}}); err != nil {
		t.Fatal(err)
	}
	state, event := read(conn), read(conn)
	if state.Sequence != baseline.Sequence+1 || event.Sequence != state.Sequence+1 || event.Type != "event" {
		t.Fatal("stream ordering mismatch")
	}
	_ = conn.CloseNow()
	reconnected, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer reconnected.CloseNow()
	recovered := read(reconnected)
	var restored telemetry.State
	if err := json.Unmarshal(recovered.Payload, &restored); err != nil {
		t.Fatal(err)
	}
	if recovered.Type != "snapshot" || recovered.Sequence != event.Sequence || restored.Brain.OutputEvents != 1 {
		t.Fatal("reconnection lost the current state")
	}
	handler.Close()
	if _, _, err := reconnected.Read(ctx); err == nil {
		t.Fatal("handler shutdown left the WebSocket open")
	}
}

func TestOriginAndMethodBoundaries(t *testing.T) {
	hub, err := telemetry.New(telemetry.State{Mode: "test"})
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Close()
	handler := New(hub)
	defer handler.Close()
	server := httptest.NewServer(handler)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, response, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(server.URL, "http")+"/api/v1/stream", &websocket.DialOptions{HTTPHeader: http.Header{"Origin": {"https://unrelated.example"}}})
	if conn != nil {
		_ = conn.CloseNow()
	}
	if err == nil || response == nil || response.StatusCode != http.StatusForbidden {
		t.Fatalf("cross-origin stream was accepted: %v", err)
	}
	for _, path := range []string{"/api/v1/snapshot", "/api/v1/stream"} {
		request, _ := http.NewRequest(http.MethodPost, server.URL+path, nil)
		response, err := server.Client().Do(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusMethodNotAllowed {
			t.Fatalf("POST %s: %d", path, response.StatusCode)
		}
	}
}
