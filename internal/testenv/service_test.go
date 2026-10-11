package testenv

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestControlAndHealth(t *testing.T) {
	service := New()
	control := httptest.NewRecorder()
	service.ServeHTTP(control, httptest.NewRequest(http.MethodPut, "/control", strings.NewReader(`{"statusCode":503,"delayMs":20}`)))
	if control.Code != http.StatusOK {
		t.Fatalf("control: %d", control.Code)
	}
	start := time.Now()
	health := httptest.NewRecorder()
	service.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/health", nil))
	if health.Code != 503 || time.Since(start) < 20*time.Millisecond {
		t.Fatalf("health: status %d, elapsed %v", health.Code, time.Since(start))
	}
	control = httptest.NewRecorder()
	service.ServeHTTP(control, httptest.NewRequest(http.MethodPut, "/control", strings.NewReader(`{"statusCode":200,"delayMs":0}`)))
	health = httptest.NewRecorder()
	service.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/health", nil))
	if health.Code != 200 {
		t.Fatalf("recovery: %d", health.Code)
	}
}

func TestInvalidControlPreservesConfiguration(t *testing.T) {
	service := New()
	for _, input := range []string{
		`{}`, `null`, `{"statusCode":199}`, `{"statusCode":600}`, `{"statusCode":200,"delayMs":-1}`,
		`{"statusCode":200,"delayMs":5001}`, `{"statusCode":200,"unknown":1}`,
		`{"statusCode":200} {}`, strings.Repeat(" ", 1025) + `{"statusCode":200}`,
	} {
		response := httptest.NewRecorder()
		service.ServeHTTP(response, httptest.NewRequest(http.MethodPut, "/control", strings.NewReader(input)))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("accepted %q: %d", input, response.Code)
		}
	}
	response := httptest.NewRecorder()
	service.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/control", nil))
	var config Config
	if err := json.Unmarshal(response.Body.Bytes(), &config); err != nil {
		t.Fatal(err)
	}
	if config != (Config{StatusCode: 200}) {
		t.Fatalf("invalid control changed state: %+v", config)
	}
}

func TestCanceledHealthDoesNotWaitForDelay(t *testing.T) {
	service := New()
	service.config.DelayMS = 5000
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := httptest.NewRequest(http.MethodGet, "/health", nil).WithContext(ctx)
	start := time.Now()
	service.ServeHTTP(httptest.NewRecorder(), request)
	if time.Since(start) > time.Second {
		t.Fatal("cancellation waited for the simulated delay")
	}
}
