package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestApplicationLifecycle(t *testing.T) {
	for _, mode := range []string{"test", "brain"} {
		t.Run(mode, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			address := listener.Addr().String()
			cfg := Config{Mode: mode, BrainPath: writeTestBrain(t, 101), TestListenAddress: "127.0.0.1:0"}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			result := make(chan error, 1)
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			go func() { result <- runWithListener(ctx, cfg, logger, listener) }()
			client := &http.Client{Timeout: 2 * time.Second}
			defer client.CloseIdleConnections()
			for _, route := range []string{"/", "/lab", "/brain"} {
				response, err := client.Get("http://" + address + route)
				if err != nil {
					t.Fatal(err)
				}
				body, err := io.ReadAll(response.Body)
				response.Body.Close()
				if err != nil || response.StatusCode != http.StatusOK || !strings.Contains(string(body), "Drosophila.OS") {
					t.Fatalf("%s returned %d: %s, %v", route, response.StatusCode, body, err)
				}
			}
			// More than three 250ms ticks: the default test runtime must still be alive.
			select {
			case err := <-result:
				t.Fatalf("application exited before cancellation: %v", err)
			case <-time.After(time.Second):
			}
			cancel()
			select {
			case err := <-result:
				if err != nil {
					t.Fatalf("shutdown: %v", err)
				}
			case <-time.After(10 * time.Second):
				t.Fatal("application did not stop")
			}
			assertAddressReleased(t, address)
		})
	}
}

func TestBackendFailureStopsHTTP(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := Config{Mode: "brain", BrainPath: filepath.Join(t.TempDir(), "missing.bin")}
	if err := runWithListener(ctx, cfg, logger, listener); err == nil {
		t.Fatal("expected backend failure")
	}
	assertAddressReleased(t, address)
}

func TestHTTPFailureStopsBackend(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	listener.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	err = runWithListener(ctx, Config{Mode: "test", TestListenAddress: "127.0.0.1:0"}, logger, listener)
	if !errors.Is(err, net.ErrClosed) {
		t.Fatalf("expected listener failure, got %v", err)
	}
}

func TestOccupiedAddressFailsBeforeBackend(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := Config{Mode: "brain", BrainPath: "missing.bin", ListenAddress: listener.Addr().String()}
	if err := Run(context.Background(), cfg, logger); err == nil || !strings.Contains(err.Error(), "listen:") {
		t.Fatalf("expected listen failure before graph loading, got %v", err)
	}
}

func assertAddressReleased(t *testing.T, address string) {
	t.Helper()
	listener, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatalf("HTTP address was not released: %v", err)
	}
	listener.Close()
}

func TestEnvironmentSharesApplicationLifetime(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	records := make(chan slog.Record, 128)
	result := make(chan error, 1)
	cfg := Config{Mode: "test", BrainPath: "unused.bin", ListenAddress: "127.0.0.1:0", TestListenAddress: "127.0.0.1:0"}
	go func() { result <- Run(ctx, cfg, slog.New(recordHandler{records})) }()
	var addresses []string
	defer func() {
		cancel()
		select {
		case err := <-result:
			if err != nil {
				t.Errorf("shutdown: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("application did not stop")
		}
		for _, address := range addresses {
			assertAddressReleased(t, address)
		}
	}()
	// Both records may arrive in either order; capture them without discarding the other.
	started := false
	for len(addresses) < 2 || !started {
		record := waitRecord(t, records, func(record slog.Record) bool {
			return record.Message == "frontend listening" || record.Message == "test environment listening" || record.Message == "test runtime started"
		})
		if record.Message == "test runtime started" {
			started = true
			continue
		}
		var address string
		record.Attrs(func(attr slog.Attr) bool {
			if attr.Key == "address" {
				address = attr.Value.String()
			}
			return true
		})
		addresses = append(addresses, address)
		path := "/"
		if record.Message == "test environment listening" {
			path = "/control"
		}
		client := &http.Client{Timeout: 2 * time.Second}
		response, err := client.Get("http://" + address + path)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		client.CloseIdleConnections()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("%s: %d", path, response.StatusCode)
		}
	}
}

func TestOccupiedEnvironmentAddressReleasesUI(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	cfg := Config{Mode: "test", TestListenAddress: occupied.Addr().String()}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := runWithListener(context.Background(), cfg, logger, listener); err == nil || !strings.Contains(err.Error(), "test environment listen:") {
		t.Fatalf("expected environment startup failure, got %v", err)
	}
	assertAddressReleased(t, address)
}
