package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/mohadayo/pulse-monitor/health-checker/checker"
	"github.com/mohadayo/pulse-monitor/health-checker/config"
	"github.com/mohadayo/pulse-monitor/health-checker/server"
)

func main() {
	cfg := config.Load()

	var level slog.Level
	switch cfg.LogLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	timeout, err := time.ParseDuration(cfg.CheckTimeout)
	if err != nil {
		logger.Error("invalid CHECK_TIMEOUT", "value", cfg.CheckTimeout, "error", err)
		os.Exit(1)
	}

	c := checker.New(timeout, logger)
	srv := server.New(c, logger)

	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.Info("starting health-checker", "addr", addr)

	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
