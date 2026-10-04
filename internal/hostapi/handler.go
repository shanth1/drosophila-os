// Package hostapi serves the versioned read-only host snapshot and event stream.
package hostapi

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/shanth1/drosophila-os/internal/telemetry"
)

type Handler struct {
	hub         *telemetry.Hub
	mux         *http.ServeMux
	mu          sync.Mutex
	connections map[*websocket.Conn]struct{}
	closed      bool
	wg          sync.WaitGroup
}

func New(hub *telemetry.Hub) *Handler {
	handler := &Handler{hub: hub, mux: http.NewServeMux(), connections: make(map[*websocket.Conn]struct{})}
	handler.mux.HandleFunc("GET /api/v1/snapshot", handler.snapshot)
	handler.mux.HandleFunc("GET /api/v1/stream", handler.stream)
	return handler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) { h.mux.ServeHTTP(w, r) }

func (h *Handler) snapshot(w http.ResponseWriter, r *http.Request) {
	data, err := h.hub.Snapshot()
	if err != nil {
		http.Error(w, "snapshot unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method != http.MethodHead {
		_, _ = w.Write(data)
	}
}

func (h *Handler) stream(w http.ResponseWriter, r *http.Request) {
	// Default origin checks are retained. Vite preserves the browser Host header.
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		_ = conn.CloseNow()
		return
	}
	h.connections[conn] = struct{}{}
	h.wg.Add(1)
	h.mu.Unlock()
	defer func() {
		_ = conn.CloseNow()
		h.mu.Lock()
		delete(h.connections, conn)
		h.mu.Unlock()
		h.wg.Done()
	}()
	ctx := conn.CloseRead(r.Context())
	subscriber, err := h.hub.Subscribe()
	if err != nil {
		return
	}
	defer h.hub.Unsubscribe(subscriber)
	if err := writeMessage(ctx, conn, subscriber.Snapshot); err != nil {
		return
	}
	heartbeat := time.NewTicker(10 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-subscriber.Done:
			_ = conn.Close(websocket.StatusTryAgainLater, "reconnect for a fresh snapshot")
			return
		case data := <-subscriber.Updates:
			if err := writeMessage(ctx, conn, data); err != nil {
				return
			}
		case <-heartbeat.C:
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			err := conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

func writeMessage(ctx context.Context, conn *websocket.Conn, data []byte) error {
	writeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return conn.Write(writeCtx, websocket.MessageText, data)
}

// Close terminates hijacked WebSockets, which http.Server.Shutdown does not own,
// and waits for all registered stream handlers. No new handlers can register.
func (h *Handler) Close() {
	h.mu.Lock()
	h.closed = true
	connections := make([]*websocket.Conn, 0, len(h.connections))
	for conn := range h.connections {
		connections = append(connections, conn)
	}
	h.mu.Unlock()
	for _, conn := range connections {
		_ = conn.CloseNow()
	}
	h.wg.Wait()
}
