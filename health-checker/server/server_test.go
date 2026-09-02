package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mohadayo/pulse-monitor/health-checker/checker"
)

func newTestServer() *Server {
	logger := slog.Default()
	c := checker.New(5*time.Second, logger)
	return New(c, logger)
}

func TestHealthEndpoint(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected ok, got %s", resp.Status)
	}
	if resp.Service != "health-checker" {
		t.Errorf("expected health-checker, got %s", resp.Service)
	}
}

func TestCheckEndpoint(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	s := newTestServer()
	body, _ := json.Marshal(CheckRequest{URL: target.URL})
	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result checker.Result
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if result.Status != "healthy" {
		t.Errorf("expected healthy, got %s", result.Status)
	}
}

func TestCheckEndpointMissingURL(t *testing.T) {
	s := newTestServer()
	body, _ := json.Marshal(CheckRequest{URL: ""})
	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewReader(body))
	w := httptest.NewRecorder()

	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCheckEndpointInvalidBody(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCheckEndpointInvalidURL(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{"missing scheme", "example.com/health"},
		{"non http scheme file", "file:///etc/passwd"},
		{"non http scheme ftp", "ftp://example.com/pub"},
		{"non http scheme gopher", "gopher://example.com"},
		{"missing host", "http://"},
		{"scheme only", "https://"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := newTestServer()
			body, _ := json.Marshal(CheckRequest{URL: tc.url})
			req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewReader(body))
			w := httptest.NewRecorder()

			s.Handler().ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("url=%q: expected 400, got %d (body=%q)", tc.url, w.Code, w.Body.String())
			}

			ct := w.Header().Get("Content-Type")
			if ct == "" {
				t.Errorf("url=%q: expected Content-Type header to be set", tc.url)
			}

			var resp map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("url=%q: expected JSON error body, got %q: %v", tc.url, w.Body.String(), err)
			}
			if _, ok := resp["error"]; !ok {
				t.Errorf("url=%q: expected 'error' key in response body, got %v", tc.url, resp)
			}
		})
	}
}

// TestCheckEndpointPropagatesClientCancellation は、クライアントが /check への
// リクエストをキャンセル (=リクエストコンテキストが Done) した場合に、
// ハンドラがそのコンテキストをアウトバウンドの HTTP チェックへ伝播させ、
// 対象サーバが応答を返す前にチェックが中断されることを検証する。
//
// これまでは context.Background() が使われていたため、クライアントが
// 切断してもアウトバウンドチェックはブロックし続けていた。
func TestCheckEndpointPropagatesClientCancellation(t *testing.T) {
	// 遅延応答するターゲット。ctx キャンセルで即座に抜ける。
	reached := make(chan struct{}, 1)
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached <- struct{}{}
		select {
		case <-r.Context().Done():
			return
		case <-time.After(5 * time.Second):
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer target.Close()

	s := newTestServer()

	// 事前にキャンセル済みの ctx をリクエストに紐付け。ハンドラが r.Context()
	// を利用していれば、アウトバウンド HTTP チェックは即座にエラー扱いになる。
	ctx, cancel := context.WithCancel(context.Background())
	body, _ := json.Marshal(CheckRequest{URL: target.URL})
	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewReader(body)).WithContext(ctx)
	w := httptest.NewRecorder()

	// リクエスト送信直後にキャンセル。ハンドラは 200 を返しつつ、result は
	// unhealthy になっているはず (context.Canceled が Do() から返る)。
	done := make(chan struct{})
	go func() {
		s.Handler().ServeHTTP(w, req)
		close(done)
	}()

	// ターゲットに到達したらキャンセル。
	select {
	case <-reached:
	case <-time.After(2 * time.Second):
		cancel()
		<-done
		t.Fatalf("target was never reached")
	}
	cancel()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("handler did not return after cancellation; context is not being propagated")
	}

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 (handler always writes JSON result), got %d", w.Code)
	}

	var result checker.Result
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to decode result: %v (body=%q)", err, w.Body.String())
	}
	if result.Status != "unhealthy" {
		t.Errorf("expected unhealthy after cancellation, got %s (error=%q)", result.Status, result.Error)
	}
	if result.Error == "" {
		t.Errorf("expected non-empty error on canceled request")
	}
}
