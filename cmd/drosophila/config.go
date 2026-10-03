package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
)

type Config struct {
	BrainPath string
	TickLimit int
}

func parseConfig(args []string, helpOutput io.Writer) (Config, error) {
	var cfg Config
	flags := flag.NewFlagSet("drosophila", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&cfg.BrainPath, "brain", "data/male_cns.bin", "path to the binary connectome")
	flags.IntVar(&cfg.TickLimit, "ticks", 300, "number of biological ticks; 0 runs until Ctrl+C")
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
	if cfg.BrainPath == "" {
		return Config{}, fmt.Errorf("brain path must not be empty")
	}
	if cfg.TickLimit < 0 {
		return Config{}, fmt.Errorf("ticks must be non-negative")
	}
	return cfg, nil
}
