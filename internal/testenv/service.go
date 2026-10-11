// Package testenv provides a controllable external HTTP system for integration work.
package testenv

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

// Config controls subsequent health requests. In-flight requests keep their snapshot.
type Config struct {
	StatusCode int `json:"statusCode"`
	DelayMS    int `json:"delayMs"`
}

// Service simulates the observed system, not the Drosophila host API.
type Service struct {
	mu     sync.RWMutex
	config Config
}

func New() *Service {
	return &Service{config: Config{StatusCode: http.StatusOK}}
}

func (s *Service) snapshot() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/health":
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		config := s.snapshot()
		timer := time.NewTimer(time.Duration(config.DelayMS) * time.Millisecond)
		defer timer.Stop()
		select {
		case <-r.Context().Done():
			return
		case <-timer.C:
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(config.StatusCode)
		if config.StatusCode != http.StatusNoContent && config.StatusCode != http.StatusNotModified {
			_, _ = io.WriteString(w, "test environment\n")
		}
	case "/control":
		switch r.Method {
		case http.MethodGet:
			writeConfig(w, s.snapshot())
		case http.MethodPut:
			decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
			decoder.DisallowUnknownFields()
			var config Config
			if err := decoder.Decode(&config); err != nil {
				http.Error(w, "invalid configuration JSON", http.StatusBadRequest)
				return
			}
			if err := decoder.Decode(new(any)); err != io.EOF {
				http.Error(w, "expected one JSON object", http.StatusBadRequest)
				return
			}
			if config.StatusCode < 200 || config.StatusCode > 599 || config.DelayMS < 0 || config.DelayMS > 5000 {
				http.Error(w, "statusCode must be 200..599; delayMs must be 0..5000", http.StatusBadRequest)
				return
			}
			s.mu.Lock()
			s.config = config
			s.mu.Unlock()
			writeConfig(w, config)
		default:
			w.Header().Set("Allow", "GET, PUT")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.NotFound(w, r)
	}
}

func writeConfig(w http.ResponseWriter, config Config) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(config)
}
