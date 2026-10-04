// Package httpmonitor performs bounded HTTP observations without interpreting health.
package httpmonitor

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Result reports response-header latency. StatusCode is zero for transport failures.
type Result struct {
	StatusCode    uint32
	ElapsedMillis uint32
	Err           error
}

// Probe owns a fixed host-approved target and its HTTP connection pool.
type Probe struct {
	target string
	client *http.Client
}

func New(target string, timeout time.Duration) (*Probe, error) {
	u, err := url.Parse(target)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("probe target must be an absolute HTTP(S) URL")
	}
	if timeout <= 0 || timeout > time.Minute {
		return nil, fmt.Errorf("probe timeout must be positive and at most one minute")
	}
	return &Probe{target: target, client: &http.Client{
		Timeout:   timeout,
		Transport: http.DefaultTransport.(*http.Transport).Clone(),
		// Preserve the observed response rather than probing an unrelated redirect target.
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}}, nil
}

func (p *Probe) Measure(ctx context.Context) Result {
	start := time.Now()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.target, nil)
	if err != nil {
		return Result{Err: err}
	}
	response, err := p.client.Do(request)
	elapsed := time.Since(start).Milliseconds()
	if elapsed > int64(^uint32(0)) {
		elapsed = int64(^uint32(0))
	}
	result := Result{ElapsedMillis: uint32(elapsed), Err: err}
	if err == nil {
		result.StatusCode = uint32(response.StatusCode)
		// The measurement ends at headers; do not wait for an arbitrary response body.
		_ = response.Body.Close()
	}
	return result
}

func (p *Probe) Close() { p.client.CloseIdleConnections() }
