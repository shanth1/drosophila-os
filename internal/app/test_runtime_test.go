package app

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shanth1/drosophila-os/internal/telemetry"
	"github.com/shanth1/drosophila-os/internal/testenv"
)

type recordHandler struct{ records chan slog.Record }

func (h recordHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h recordHandler) Handle(_ context.Context, record slog.Record) error {
	select {
	case h.records <- record.Clone():
	default:
	}
	return nil
}
func (h recordHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h recordHandler) WithGroup(string) slog.Handler      { return h }

func waitRecord(t *testing.T, records <-chan slog.Record, matches func(slog.Record) bool) slog.Record {
	t.Helper()
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	for {
		select {
		case record := <-records:
			if matches(record) {
				return record
			}
		case <-timer.C:
			t.Fatal("expected runtime record did not arrive")
			return slog.Record{}
		}
	}
}

func observedStatus(status int64) func(slog.Record) bool {
	return func(record slog.Record) bool {
		if record.Message != "HTTP observation" {
			return false
		}
		matched := false
		record.Attrs(func(attr slog.Attr) bool {
			if attr.Key == "status" && attr.Value.Uint64() == uint64(status) {
				matched = true
			}
			return true
		})
		return matched
	}
}

func TestHTTPPipelineFailureAndRecovery(t *testing.T) {
	service := httptest.NewServer(testenv.New())
	defer service.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	records := make(chan slog.Record, 128)
	result := make(chan error, 1)
	hub := newTestHub(t)
	subscriber, err := hub.Subscribe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { hub.Unsubscribe(subscriber) })
	go func() {
		result <- runTestRuntime(ctx, Config{}, slog.New(recordHandler{records}), service.URL+"/health", hub)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-result:
			if err != nil {
				t.Errorf("runtime shutdown: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("runtime did not stop")
		}
	})
	waitRecord(t, records, observedStatus(200))
	setStatus := func(status string) {
		t.Helper()
		request, err := http.NewRequest(http.MethodPut, service.URL+"/control", strings.NewReader(`{"statusCode":`+status+`,"delayMs":0}`))
		if err != nil {
			t.Fatal(err)
		}
		response, err := service.Client().Do(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("control returned %d", response.StatusCode)
		}
	}
	setStatus("503")
	waitRecord(t, records, observedStatus(503))
	waitRecord(t, records, func(record slog.Record) bool { return record.Message == "output event" })
	setStatus("200")
	waitRecord(t, records, observedStatus(200))
	// Allow previously scheduled propagation to drain before checking recovery.
	settle := time.NewTimer(time.Second)
	defer settle.Stop()
	settling := true
	for settling {
		select {
		case <-records:
		case <-settle.C:
			settling = false
		}
	}
	window := time.NewTimer(time.Second)
	defer window.Stop()
	for {
		select {
		case record := <-records:
			if record.Message == "output event" {
				t.Fatal("output continued after healthy recovery")
			}
		case <-window.C:
			data, err := hub.Snapshot()
			if err != nil {
				t.Fatal(err)
			}
			var snapshot telemetry.Message
			if err := json.Unmarshal(data, &snapshot); err != nil {
				t.Fatal(err)
			}
			var state telemetry.State
			if err := json.Unmarshal(snapshot.Payload, &state); err != nil {
				t.Fatal(err)
			}
			if state.Brain.OutputEvents == 0 || state.Brain.LastOutputAt == nil || len(state.Observations) != 1 {
				t.Fatalf("runtime snapshot omitted output or observation state: %+v", state)
			}
			measurement, err := json.Marshal(state.Observations[0].Data)
			if err != nil {
				t.Fatal(err)
			}
			var response telemetry.HTTPResponse
			if err := json.Unmarshal(measurement, &response); err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != 200 {
				t.Fatalf("snapshot did not reflect HTTP recovery: %+v", response)
			}
			found := false
			for len(subscriber.Updates) > 0 {
				var message telemetry.Message
				if err := json.Unmarshal(<-subscriber.Updates, &message); err != nil {
					t.Fatal(err)
				}
				if message.Type == "event" {
					var event telemetry.Event
					if err := json.Unmarshal(message.Payload, &event); err != nil {
						t.Fatal(err)
					}
					if event.Name == "neural.output.spike" && event.Subject == "test-network" {
						found = true
					}
				}
			}
			if !found {
				t.Fatal("runtime did not publish output events to the stream")
			}
			return
		}
	}
}

func TestSlowHTTPDoesNotBlockEngineTicks(t *testing.T) {
	service := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer service.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var output bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug}))
	if err := runTestRuntime(ctx, Config{TickLimit: 3}, logger, service.URL, newTestHub(t)); err != nil {
		t.Fatal(err)
	}
	if states := strings.Count(output.String(), `msg="test runtime state"`); states != 3 {
		t.Fatalf("expected 3 independent engine ticks, got %d", states)
	}
	if strings.Contains(output.String(), "HTTP observation") {
		t.Fatal("the blocked HTTP request unexpectedly completed before the tick limit")
	}
}

func newTestHub(t *testing.T) *telemetry.Hub {
	t.Helper()
	hub, err := telemetry.New(telemetry.State{Mode: "test", Status: "starting"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(hub.Close)
	return hub
}
