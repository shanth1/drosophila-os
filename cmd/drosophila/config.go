package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

type Config struct {
	Mode          string
	BrainPath     string
	TickLimit     int
	ListenAddress string
}

func parseConfig(args []string, helpOutput io.Writer) (Config, error) {
	var cfg Config
	flags := flag.NewFlagSet("drosophila", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&cfg.Mode, "mode", "test", "backend mode: test (fixed sensor) or brain (experimental graph)")
	flags.StringVar(&cfg.ListenAddress, "listen", "127.0.0.1:8080", "HTTP listen address (UI is always served)")
	flags.StringVar(&cfg.BrainPath, "brain", "data/male_cns.bin", "path to the binary connectome (brain mode only)")
	flags.IntVar(&cfg.TickLimit, "ticks", 0, "optional biological tick limit; 0 runs until Ctrl+C")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			flags.SetOutput(helpOutput)
			flags.Usage()
		}
		return Config{}, err
	}
	if flags.NArg() != 0 {
		return Config{}, fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	if cfg.Mode != "test" && cfg.Mode != "brain" {
		return Config{}, fmt.Errorf("mode must be test or brain")
	}
	if cfg.ListenAddress == "" {
		return Config{}, fmt.Errorf("listen address must not be empty")
	}
	if cfg.BrainPath == "" {
		return Config{}, fmt.Errorf("brain path must not be empty")
	}
	if cfg.TickLimit < 0 {
		return Config{}, fmt.Errorf("ticks must be non-negative")
	}
	return cfg, nil
}
