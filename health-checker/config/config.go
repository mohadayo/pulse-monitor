package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port              string
	LogLevel          string
	APIGatewayURL     string
	AlertServiceURL   string
	CheckTimeout      string
	ReadHeaderTimeout string
	ReadTimeout       string
	WriteTimeout      string
	IdleTimeout       string
	ShutdownTimeout   string
}

// defaultLogLevel is what `resolveLogLevel` falls back to for empty or
// unrecognized LOG_LEVEL input; it matches `main.go`'s switch default and
// the `api-gateway` / `alert-service` config defaults.
const defaultLogLevel = "INFO"

// allowedLogLevels is the set of normalized LOG_LEVEL values `main.go`'s
// switch dispatches on. Keep in sync with that switch.
var allowedLogLevels = map[string]struct{}{
	"DEBUG": {}, "INFO": {}, "WARN": {}, "ERROR": {},
}

func Load() *Config {
	return &Config{
		Port:              getEnv("CHECKER_PORT", "8001"),
		LogLevel:          resolveLogLevel(os.Getenv("LOG_LEVEL")),
		APIGatewayURL:     getEnv("API_GATEWAY_URL", "http://api-gateway:8000"),
		AlertServiceURL:   getEnv("ALERT_SERVICE_URL", "http://alert-service:8002"),
		CheckTimeout:      getEnv("CHECK_TIMEOUT", "5s"),
		ReadHeaderTimeout: getEnv("READ_HEADER_TIMEOUT", "5s"),
		ReadTimeout:       getEnv("READ_TIMEOUT", "15s"),
		WriteTimeout:      getEnv("WRITE_TIMEOUT", "15s"),
		IdleTimeout:       getEnv("IDLE_TIMEOUT", "60s"),
		ShutdownTimeout:   getEnv("SHUTDOWN_TIMEOUT", "30s"),
	}
}

// resolveLogLevel normalizes LOG_LEVEL to one of the four strings
// `main.go` dispatches on (DEBUG / INFO / WARN / ERROR). It mirrors
// `api-gateway/app/config.py` and `alert-service/src/config.ts`:
// whitespace is trimmed and the value is upper-cased, "WARNING" is an
// alias for "WARN" (Python's canonical name), and unrecognized values
// fall back to INFO with a one-line warning on stderr so the
// misconfiguration surfaces in container logs. The process is never
// killed for a bad log level. Previously `main.go`'s switch was the only
// gate, so `debug`, `WARNING`, and `DEBUG ` silently resolved to INFO.
func resolveLogLevel(raw string) string {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	if normalized == "" {
		return defaultLogLevel
	}
	if normalized == "WARNING" {
		return "WARN"
	}
	if _, ok := allowedLogLevels[normalized]; ok {
		return normalized
	}
	fmt.Fprintf(os.Stderr,
		"[config] WARNING: invalid LOG_LEVEL=%q; falling back to %s. "+
			"Allowed values: DEBUG, INFO, WARN (alias: WARNING), ERROR.\n",
		raw, defaultLogLevel)
	return defaultLogLevel
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
