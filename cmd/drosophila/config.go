package main

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/shanth1/drosophila-os/internal/app"
)

func parseConfig(args []string, helpOutput io.Writer) (app.Config, error) {
	var cfg app.Config
	flags := flag.NewFlagSet("drosophila", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&cfg.Mode, "mode", "test", "backend mode: test (controlled HTTP service) or brain (experimental graph)")
	flags.StringVar(&cfg.ListenAddress, "listen", "127.0.0.1:8080", "HTTP listen address (UI is always served)")
	flags.StringVar(&cfg.TestListenAddress, "test-listen", "127.0.0.1:8081", "controlled HTTP service listen address (test mode only)")
	flags.StringVar(&cfg.BrainPath, "brain", "data/male_cns.bin", "path to the binary connectome (brain mode only)")
	flags.IntVar(&cfg.TickLimit, "ticks", 0, "optional biological tick limit; 0 runs until Ctrl+C")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			flags.SetOutput(helpOutput)
			flags.Usage()
		}
		return app.Config{}, err
	}
	if flags.NArg() != 0 {
		return app.Config{}, fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	if err := cfg.Validate(); err != nil {
		return app.Config{}, err
	}
	return cfg, nil
}
