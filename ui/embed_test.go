package ui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFrontendRoutes(t *testing.T) {
	handler := Handler()
	for _, path := range []string{"/", "/brain", "/lab"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "Drosophila.OS") {
				t.Fatalf("route returned %d: %s", response.Code, response.Body.String())
			}
		})
	}
	for _, test := range []struct {
		method, path string
		status       int
	}{
		{http.MethodGet, "/assets/missing.js", http.StatusNotFound},
		{http.MethodGet, "/api/missing", http.StatusNotFound},
		{http.MethodPost, "/", http.StatusMethodNotAllowed},
		{http.MethodHead, "/lab", http.StatusOK},
	} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(test.method, test.path, nil))
		if response.Code != test.status {
			t.Fatalf("%s %s returned %d, want %d", test.method, test.path, response.Code, test.status)
		}
		if test.method == http.MethodHead && response.Body.Len() != 0 {
			t.Fatal("HEAD returned a response body")
		}
	}
}
