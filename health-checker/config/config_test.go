package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Clear all env vars the loader consults so we exercise the defaults.
	for _, k := range []string{
		"CHECKER_PORT",
		"LOG_LEVEL",
		"API_GATEWAY_URL",
		"ALERT_SERVICE_URL",
		"CHECK_TIMEOUT",
		"READ_HEADER_TIMEOUT",
		"READ_TIMEOUT",
		"WRITE_TIMEOUT",
		"IDLE_TIMEOUT",
		"SHUTDOWN_TIMEOUT",
	} {
		t.Setenv(k, "")
	}

	cfg := Load()

	cases := map[string]string{
		"Port":              cfg.Port,
		"LogLevel":          cfg.LogLevel,
		"APIGatewayURL":     cfg.APIGatewayURL,
		"AlertServiceURL":   cfg.AlertServiceURL,
		"CheckTimeout":      cfg.CheckTimeout,
		"ReadHeaderTimeout": cfg.ReadHeaderTimeout,
		"ReadTimeout":       cfg.ReadTimeout,
		"WriteTimeout":      cfg.WriteTimeout,
		"IdleTimeout":       cfg.IdleTimeout,
		"ShutdownTimeout":   cfg.ShutdownTimeout,
	}
	want := map[string]string{
		"Port":              "8001",
		"LogLevel":          "INFO",
		"APIGatewayURL":     "http://api-gateway:8000",
		"AlertServiceURL":   "http://alert-service:8002",
		"CheckTimeout":      "5s",
		"ReadHeaderTimeout": "5s",
		"ReadTimeout":       "15s",
		"WriteTimeout":      "15s",
		"IdleTimeout":       "60s",
		"ShutdownTimeout":   "30s",
	}

	for field, got := range cases {
		if got != want[field] {
			t.Errorf("Load() %s = %q, want %q", field, got, want[field])
		}
	}
}

func TestLoadOverridesFromEnv(t *testing.T) {
	overrides := map[string]string{
		"CHECKER_PORT":        "9001",
		"LOG_LEVEL":           "DEBUG",
		"API_GATEWAY_URL":     "http://gw:1",
		"ALERT_SERVICE_URL":   "http://alerts:2",
		"CHECK_TIMEOUT":       "1s",
		"READ_HEADER_TIMEOUT": "2s",
		"READ_TIMEOUT":        "3s",
		"WRITE_TIMEOUT":       "4s",
		"IDLE_TIMEOUT":        "5s",
		"SHUTDOWN_TIMEOUT":    "6s",
	}
	for k, v := range overrides {
		t.Setenv(k, v)
	}

	cfg := Load()

	if cfg.Port != "9001" {
		t.Errorf("Port = %q, want 9001", cfg.Port)
	}
	if cfg.LogLevel != "DEBUG" {
		t.Errorf("LogLevel = %q, want DEBUG", cfg.LogLevel)
	}
	if cfg.APIGatewayURL != "http://gw:1" {
		t.Errorf("APIGatewayURL = %q", cfg.APIGatewayURL)
	}
	if cfg.AlertServiceURL != "http://alerts:2" {
		t.Errorf("AlertServiceURL = %q", cfg.AlertServiceURL)
	}
	if cfg.CheckTimeout != "1s" {
		t.Errorf("CheckTimeout = %q", cfg.CheckTimeout)
	}
	if cfg.ReadHeaderTimeout != "2s" {
		t.Errorf("ReadHeaderTimeout = %q", cfg.ReadHeaderTimeout)
	}
	if cfg.ReadTimeout != "3s" {
		t.Errorf("ReadTimeout = %q", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != "4s" {
		t.Errorf("WriteTimeout = %q", cfg.WriteTimeout)
	}
	if cfg.IdleTimeout != "5s" {
		t.Errorf("IdleTimeout = %q", cfg.IdleTimeout)
	}
	if cfg.ShutdownTimeout != "6s" {
		t.Errorf("ShutdownTimeout = %q", cfg.ShutdownTimeout)
	}
}
