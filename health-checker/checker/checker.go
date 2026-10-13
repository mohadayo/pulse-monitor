package checker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// Result is the JSON payload returned for a single health check.
//
// LatencyMS is measured in milliseconds so that consumers (and the
// documented API contract) match the on-the-wire field name
// `latency_ms`. Previously this field was a time.Duration, which the
// standard encoder emits as its underlying int64 nanosecond count and
// therefore returned values three orders of magnitude larger than the
// field name promised.
type Result struct {
	URL       string    `json:"url"`
	Status    string    `json:"status"`
	Code      int       `json:"status_code"`
	LatencyMS int64     `json:"latency_ms"`
	Error     string    `json:"error,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}

type Checker struct {
	client  *http.Client
	timeout time.Duration
	logger  *slog.Logger
}

// drainAndClose は HTTP レスポンスボディを全読み込みしてから Close する。
//
// net/http.Transport は Body が読み切られていない接続を keep-alive プールに
// 戻さず破棄する。/health のような小さなレスポンスでも Close() のみでは
// 接続が破棄され、health-checker のポーリング（複数ターゲット × interval）
// で毎回 TCP+TLS ハンドシェイクが張り直されるコストが積み上がる。
//
// io.Discard への io.Copy は内部で小さな既読バッファを再利用するため
// 割り当てオーバーヘッドはない。Close のエラーは呼び元が対処できないため無視する。
func drainAndClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, body)
	_ = body.Close()
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
			LatencyMS: time.Since(start).Milliseconds(),
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
			LatencyMS: latency.Milliseconds(),
			Error:     fmt.Sprintf("request failed: %v", err),
			CheckedAt: time.Now().UTC(),
		}
	}
	defer drainAndClose(resp.Body)

	status := "healthy"
	if resp.StatusCode >= 400 {
		status = "unhealthy"
	}

	c.logger.Info("health check completed", "url", url, "status", status, "code", resp.StatusCode, "latency", latency)
	return Result{
		URL:       url,
		Status:    status,
		Code:      resp.StatusCode,
		LatencyMS: latency.Milliseconds(),
		CheckedAt: time.Now().UTC(),
	}
}
