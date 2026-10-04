package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"strings"
	"testing"

	"github.com/shanth1/drosophila-os/internal/app"
)

func TestParseConfig(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want app.Config
	}{
		{"defaults", nil, app.Config{Mode: "test", BrainPath: "data/male_cns.bin", TickLimit: 0, ListenAddress: "127.0.0.1:8080"}},
		{"overrides", []string{"-mode", "brain", "-brain", "custom.bin", "-ticks", "10"}, app.Config{Mode: "brain", BrainPath: "custom.bin", TickLimit: 10, ListenAddress: "127.0.0.1:8080"}},
		{"unlimited", []string{"-ticks", "0"}, app.Config{Mode: "test", BrainPath: "data/male_cns.bin", TickLimit: 0, ListenAddress: "127.0.0.1:8080"}},
		{"test address", []string{"-mode", "test", "-listen", "127.0.0.1:8090"}, app.Config{Mode: "test", BrainPath: "data/male_cns.bin", TickLimit: 0, ListenAddress: "127.0.0.1:8090"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseConfig(tc.args, io.Discard)
			if err != nil || got != tc.want {
				t.Fatalf("got %+v, %v; want %+v", got, err, tc.want)
			}
		})
	}
}

func TestParseConfigRejectsInvalidArguments(t *testing.T) {
	for _, args := range [][]string{
		{"-ticks", "-1"}, {"-ticks", "invalid"}, {"-brain", ""},
		{"-unknown"}, {"-brain"}, {"unexpected"}, {"-mode", "unknown"},
		{"-mode", "demo"}, {"-mode", "ui"}, {"-listen", ""},
	} {
		var output bytes.Buffer
		if _, err := parseConfig(args, &output); err == nil {
			t.Fatalf("accepted invalid arguments: %v", args)
		}
		if output.Len() != 0 {
			t.Fatalf("parser must return errors without logging them: %s", output.String())
		}
	}
}

func TestParseConfigHelp(t *testing.T) {
	var output bytes.Buffer
	_, err := parseConfig([]string{"-help"}, &output)
	if !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("expected help, got %v", err)
	}
	for _, text := range []string{"Usage of drosophila", "-mode", "-brain", "-ticks", "-listen", `default "test"`, "0 runs until Ctrl+C"} {
		if !strings.Contains(output.String(), text) {
			t.Fatalf("help missing %q: %s", text, output.String())
		}
	}
}
