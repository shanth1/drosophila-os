package main

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
			cfg := Config{Mode: mode, BrainPath: writeTestBrain(t, 101)}
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
	err = runWithListener(ctx, Config{Mode: "test"}, logger, listener)
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
	if err := run(context.Background(), cfg, logger); err == nil || !strings.Contains(err.Error(), "listen:") {
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
