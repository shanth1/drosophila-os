package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want Config
	}{
		{"defaults", nil, Config{Mode: "demo", BrainPath: "data/male_cns.bin", TickLimit: 3}},
		{"overrides", []string{"-mode", "brain", "-brain", "custom.bin", "-ticks", "10"}, Config{Mode: "brain", BrainPath: "custom.bin", TickLimit: 10}},
		{"unlimited", []string{"-ticks", "0"}, Config{Mode: "demo", BrainPath: "data/male_cns.bin", TickLimit: 0}},
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
	for _, text := range []string{"Usage of drosophila", "-mode", "-brain", "-ticks", "default 3"} {
		if !strings.Contains(output.String(), text) {
			t.Fatalf("help missing %q: %s", text, output.String())
		}
	}
}
