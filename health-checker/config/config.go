package config

import "os"

type Config struct {
	Port            string
	LogLevel        string
	APIGatewayURL   string
	AlertServiceURL string
	CheckTimeout    string
}

func Load() *Config {
	return &Config{
		Port:            getEnv("CHECKER_PORT", "8001"),
		LogLevel:        getEnv("LOG_LEVEL", "INFO"),
		APIGatewayURL:   getEnv("API_GATEWAY_URL", "http://api-gateway:8000"),
		AlertServiceURL: getEnv("ALERT_SERVICE_URL", "http://alert-service:8002"),
		CheckTimeout:    getEnv("CHECK_TIMEOUT", "5s"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
