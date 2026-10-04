package app

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
)

func TestRunRejectsInvalidConfiguration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	for _, test := range []struct {
		name   string
		change func(*Config)
		want   string
	}{
		{"mode", func(cfg *Config) { cfg.Mode = "unknown" }, "mode must be"},
		{"listen", func(cfg *Config) { cfg.ListenAddress = "" }, "listen address"},
		{"brain", func(cfg *Config) { cfg.BrainPath = "" }, "brain path"},
		{"ticks", func(cfg *Config) { cfg.TickLimit = -1 }, "ticks must be"},
		{"test listen", func(cfg *Config) { cfg.TestListenAddress = "" }, "test listen address"},
	} {
		t.Run(test.name, func(t *testing.T) {
			// An invalid network address proves validation precedes HTTP startup.
			cfg := Config{Mode: "test", BrainPath: "unused.bin", ListenAddress: "invalid-address", TestListenAddress: "127.0.0.1:0"}
			test.change(&cfg)
			err := Run(context.Background(), cfg, logger)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected configuration error %q, got %v", test.want, err)
			}
		})
	}
}
