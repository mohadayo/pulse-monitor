package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	checkTimeout := mustParseDuration(logger, "CHECK_TIMEOUT", cfg.CheckTimeout)
	readHeaderTimeout := mustParseDuration(logger, "READ_HEADER_TIMEOUT", cfg.ReadHeaderTimeout)
	readTimeout := mustParseDuration(logger, "READ_TIMEOUT", cfg.ReadTimeout)
	writeTimeout := mustParseDuration(logger, "WRITE_TIMEOUT", cfg.WriteTimeout)
	idleTimeout := mustParseDuration(logger, "IDLE_TIMEOUT", cfg.IdleTimeout)
	shutdownTimeout := mustParseDuration(logger, "SHUTDOWN_TIMEOUT", cfg.ShutdownTimeout)

	c := checker.New(checkTimeout, logger)
	srv := server.New(c, logger)

	addr := fmt.Sprintf(":%s", cfg.Port)
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	// Trap SIGINT / SIGTERM so we can drain in-flight requests before exit.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("starting health-checker", "addr", addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.Info("shutting down", "timeout", shutdownTimeout.String())
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			os.Exit(1)
		}
		logger.Info("shutdown complete")
	}
}

// mustParseDuration parses a duration string and exits the process with a
// descriptive log line on failure. It is only used at startup for values that
// come from configuration, so failing fast is the right behavior.
func mustParseDuration(logger *slog.Logger, name, value string) time.Duration {
	d, err := time.ParseDuration(value)
	if err != nil {
		logger.Error("invalid duration", "env", name, "value", value, "error", err)
		os.Exit(1)
	}
	return d
}
