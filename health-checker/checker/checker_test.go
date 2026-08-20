package checker

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckHealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	c := New(5*time.Second, slog.Default())
	result := c.Check(context.Background(), srv.URL)

	if result.Status != "healthy" {
		t.Errorf("expected healthy, got %s", result.Status)
	}
	if result.Code != 200 {
		t.Errorf("expected 200, got %d", result.Code)
	}
	if result.Error != "" {
		t.Errorf("expected no error, got %s", result.Error)
	}
	if result.LatencyMS < 0 {
		t.Errorf("expected non-negative latency in ms, got %d", result.LatencyMS)
	}
}

func TestCheckUnhealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := New(5*time.Second, slog.Default())
	result := c.Check(context.Background(), srv.URL)

	if result.Status != "unhealthy" {
		t.Errorf("expected unhealthy, got %s", result.Status)
	}
	if result.Code != 503 {
		t.Errorf("expected 503, got %d", result.Code)
	}
}

func TestCheckConnectionRefused(t *testing.T) {
	c := New(2*time.Second, slog.Default())
	result := c.Check(context.Background(), "http://127.0.0.1:19999/health")

	if result.Status != "unhealthy" {
		t.Errorf("expected unhealthy, got %s", result.Status)
	}
	if result.Error == "" {
		t.Error("expected error message")
	}
}

func TestCheckInvalidURL(t *testing.T) {
	c := New(2*time.Second, slog.Default())
	result := c.Check(context.Background(), "://invalid")

	if result.Status != "unhealthy" {
		t.Errorf("expected unhealthy, got %s", result.Status)
	}
	if result.Error == "" {
		t.Error("expected error message")
	}
}

func TestCheckTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(500*time.Millisecond, slog.Default())
	result := c.Check(context.Background(), srv.URL)

	if result.Status != "unhealthy" {
		t.Errorf("expected unhealthy due to timeout, got %s", result.Status)
	}
}

// TestResultLatencyJSONIsMilliseconds guards the contract: the wire field
// `latency_ms` must serialize as a millisecond integer, not a duration in
// nanoseconds. Regressing to time.Duration would inflate the value by
// roughly 1e6, which is exactly the bug this test protects against.
func TestResultLatencyJSONIsMilliseconds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := New(5*time.Second, slog.Default())
	result := c.Check(context.Background(), srv.URL)

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded struct {
		LatencyMS int64 `json:"latency_ms"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// 20ms server sleep; anything under ~2s is well within a millisecond
	// scale. A nanosecond-encoded value would be at least 20_000_000.
	if decoded.LatencyMS < 0 || decoded.LatencyMS > 2000 {
		t.Errorf("latency_ms=%d looks off; expected small millisecond value (not nanoseconds)", decoded.LatencyMS)
	}
}
