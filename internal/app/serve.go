package app

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"
)

func serveHTTP(ctx context.Context, listener net.Listener, handler http.Handler, logger *slog.Logger, message string) error {
	server := &http.Server{
		Handler: handler, ReadHeaderTimeout: 5 * time.Second,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}
	result := make(chan error, 1)
	go func() { result <- server.Serve(listener) }()
	logger.Info(message, "address", listener.Addr().String())
	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownErr := server.Shutdown(shutdownCtx)
		if shutdownErr != nil {
			_ = server.Close()
		}
		serveErr := <-result
		if errors.Is(serveErr, http.ErrServerClosed) {
			serveErr = nil
		}
		return errors.Join(shutdownErr, serveErr)
	}
}
