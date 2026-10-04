package app

import "fmt"

// Config describes the runtime without imposing a particular configuration source.
type Config struct {
	Mode          string
	BrainPath     string
	TickLimit     int
	ListenAddress string
}

// Validate checks runtime settings before any components are started.
func (cfg Config) Validate() error {
	if cfg.Mode != "test" && cfg.Mode != "brain" {
		return fmt.Errorf("mode must be test or brain")
	}
	if cfg.ListenAddress == "" {
		return fmt.Errorf("listen address must not be empty")
	}
	if cfg.BrainPath == "" {
		return fmt.Errorf("brain path must not be empty")
	}
	if cfg.TickLimit < 0 {
		return fmt.Errorf("ticks must be non-negative")
	}
	return nil
}
