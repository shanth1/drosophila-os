package httpmonitor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMeasurePreservesStatusAndLatency(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.Header().Set("Location", "/redirect-target")
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()
	probe, err := New(server.URL, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer probe.Close()
	result := probe.Measure(context.Background())
	if result.Err != nil || result.StatusCode != http.StatusFound || result.ElapsedMillis < 20 {
		t.Fatalf("unexpected observation: %+v", result)
	}
}

func TestTimeoutAndCancellationAreTransportFailures(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	probe, err := New(server.URL, 40*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	defer probe.Close()
	result := probe.Measure(context.Background())
	if result.Err == nil || result.StatusCode != 0 || result.ElapsedMillis < 30 {
		t.Fatalf("timeout: %+v", result)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result = probe.Measure(ctx)
	if result.Err == nil || result.StatusCode != 0 {
		t.Fatalf("cancellation: %+v", result)
	}
}
