package checker

import (
	"context"
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
	if result.Latency <= 0 {
		t.Error("expected positive latency")
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
