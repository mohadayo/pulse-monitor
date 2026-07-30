package config

import "os"

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

func Load() *Config {
	return &Config{
		Port:              getEnv("CHECKER_PORT", "8001"),
		LogLevel:          getEnv("LOG_LEVEL", "INFO"),
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

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
