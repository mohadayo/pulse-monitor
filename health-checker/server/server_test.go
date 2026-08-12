package server

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mohadayo/pulse-monitor/health-checker/checker"
)

func newTestServer() *Server {
	logger := slog.Default()
	c := checker.New(5*time.Second, logger)
	return New(c, logger)
}

func TestHealthEndpoint(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected ok, got %s", resp.Status)
	}
	if resp.Service != "health-checker" {
		t.Errorf("expected health-checker, got %s", resp.Service)
	}
}

func TestCheckEndpoint(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	s := newTestServer()
	body, _ := json.Marshal(CheckRequest{URL: target.URL})
	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result checker.Result
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if result.Status != "healthy" {
		t.Errorf("expected healthy, got %s", result.Status)
	}
}

func TestCheckEndpointMissingURL(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(CheckRequest{URL: ""})
	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCheckEndpointInvalidBody(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCheckEndpointInvalidURL(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{"missing scheme", "example.com/health"},
		{"non http scheme file", "file:///etc/passwd"},
		{"non http scheme ftp", "ftp://example.com/pub"},
		{"non http scheme gopher", "gopher://example.com"},
		{"missing host", "http://"},
		{"scheme only", "https://"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestServer()
			body, _ := json.Marshal(CheckRequest{URL: tc.url})
			req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewReader(body))
			w := httptest.NewRecorder()

			s.Handler().ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("url=%q: expected 400, got %d (body=%q)", tc.url, w.Code, w.Body.String())
			}

			ct := w.Header().Get("Content-Type")
			if ct == "" {
				t.Errorf("url=%q: expected Content-Type header to be set", tc.url)
			}

			var resp map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("url=%q: expected JSON error body, got %q: %v", tc.url, w.Body.String(), err)
			}
			if _, ok := resp["error"]; !ok {
				t.Errorf("url=%q: expected 'error' key in response body, got %v", tc.url, resp)
			}
		})
	}
}
