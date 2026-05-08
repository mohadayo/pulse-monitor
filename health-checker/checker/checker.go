package checker

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Result struct {
	URL       string        `json:"url"`
	Status    string        `json:"status"`
	Code      int           `json:"status_code"`
	Latency   time.Duration `json:"latency_ms"`
	Error     string        `json:"error,omitempty"`
	CheckedAt time.Time     `json:"checked_at"`
}

type Checker struct {
	client  *http.Client
	timeout time.Duration
	logger  *slog.Logger
}

func New(timeout time.Duration, logger *slog.Logger) *Checker {
	return &Checker{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
		logger:  logger,
	}
}

func (c *Checker) Check(ctx context.Context, url string) Result {
	c.logger.Info("checking service", "url", url)
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.logger.Error("failed to create request", "url", url, "error", err)
		return Result{
			URL:       url,
			Status:    "unhealthy",
			Code:      0,
			Latency:   time.Since(start),
			Error:     fmt.Sprintf("request creation failed: %v", err),
			CheckedAt: time.Now().UTC(),
		}
	}

	resp, err := c.client.Do(req)
	latency := time.Since(start)

	if err != nil {
		c.logger.Error("health check failed", "url", url, "error", err, "latency", latency)
		return Result{
			URL:       url,
			Status:    "unhealthy",
			Code:      0,
			Latency:   latency,
			Error:     fmt.Sprintf("request failed: %v", err),
			CheckedAt: time.Now().UTC(),
		}
	}
	defer resp.Body.Close()

	status := "healthy"
	if resp.StatusCode >= 400 {
		status = "unhealthy"
	}

	c.logger.Info("health check completed", "url", url, "status", status, "code", resp.StatusCode, "latency", latency)
	return Result{
		URL:       url,
		Status:    status,
		Code:      resp.StatusCode,
		Latency:   latency,
		CheckedAt: time.Now().UTC(),
	}
}
