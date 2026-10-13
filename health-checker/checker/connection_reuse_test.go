package checker

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// TestCheck_ReusesTCPConnection は Checker が /health のような小さな JSON
// レスポンスでも Body を drain した上で Close し、net/http.Transport の
// keep-alive プールへ接続を返却できていることを回帰する。
//
// Body が読み切られていない接続は keep-alive プールに戻らず破棄されるため、
// 連続して同じホストにチェックを行うたびに新しい TCP (および HTTPS なら TLS)
// 接続が張られる。httptest.Server.Config.ConnState で http.StateNew の
// 発生回数を数え、連続 3 回のチェックで接続が 1 本しか張られないことを確認する。
func TestCheck_ReusesTCPConnection(t *testing.T) {
	var totalConns int32
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"status":"healthy"}`)
	}))
	srv.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			atomic.AddInt32(&totalConns, 1)
		}
	}
	srv.Start()
	defer srv.Close()

	c := New(5*time.Second, slog.Default())
	for i := 0; i < 3; i++ {
		result := c.Check(context.Background(), srv.URL)
		if result.Status != "healthy" {
			t.Fatalf("iteration %d: expected healthy, got %s (error=%q)", i, result.Status, result.Error)
		}
	}

	if got := atomic.LoadInt32(&totalConns); got != 1 {
		t.Errorf("expected TCP connection reuse (totalConns=1 after 3 checks), got %d", got)
	}
}

// countingReadCloser は Read で読み込んだバイト数と Close 済みフラグを記録する
// テスト用 io.ReadCloser。drainAndClose がボディを「全部読み切ってから閉じる」
// ことを単体で検証するために使用する。
type countingReadCloser struct {
	reader *strings.Reader
	read   int
	closed bool
}

func (c *countingReadCloser) Read(p []byte) (int, error) {
	n, err := c.reader.Read(p)
	c.read += n
	return n, err
}

func (c *countingReadCloser) Close() error {
	c.closed = true
	return nil
}

// TestDrainAndClose_ReadsAllBytes は drainAndClose が渡された body の
// 全バイトを読み切り、かつ Close を呼ぶことを検証する。回帰ガード:
// 将来誰かが io.Copy を外して Close() だけに戻すと、このテストが落ちる。
func TestDrainAndClose_ReadsAllBytes(t *testing.T) {
	payload := strings.Repeat("x", 4096)
	body := &countingReadCloser{reader: strings.NewReader(payload)}
	drainAndClose(body)
	if !body.closed {
		t.Error("drainAndClose must close the body")
	}
	if body.read != len(payload) {
		t.Errorf("drainAndClose read %d bytes, want %d", body.read, len(payload))
	}
}
