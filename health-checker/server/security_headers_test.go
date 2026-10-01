package server

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestSecurityHeadersOnHealth は正常系 (200) の /health 応答に
// すべてのセキュリティヘッダが載っていることを確認する。
func TestSecurityHeadersOnHealth(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	for name, want := range SecurityHeaders {
		if got := w.Header().Get(name); got != want {
			t.Errorf("header %s: expected %q, got %q", name, want, got)
		}
	}
}

// TestSecurityHeadersOn404 はルーティング未定義パスの 404 応答にも
// ヘッダが載ることを確認する。chi の NotFound 経路にもミドルウェアが
// 効くことの回帰ガード。
func TestSecurityHeadersOn404(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	for name, want := range SecurityHeaders {
		if got := w.Header().Get(name); got != want {
			t.Errorf("header %s on 404: expected %q, got %q", name, want, got)
		}
	}
}

// TestSecurityHeadersOnBadRequest は不正な JSON ボディで 400 を返す経路でも
// ヘッダが載ることを確認する。JSON パース失敗ハンドラ (`http.Error`) 側で
// Content-Type だけ上書きしても、ミドルウェア側のセキュリティ系ヘッダは
// 残ることを確認する。
func TestSecurityHeadersOnBadRequest(t *testing.T) {
	s := newTestServer()

	req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	for name, want := range SecurityHeaders {
		if got := w.Header().Get(name); got != want {
			t.Errorf("header %s on 400: expected %q, got %q", name, want, got)
		}
	}
}

// TestSecurityHeadersValues はヘッダの具体的な値が
// `api-gateway` / `alert-service` と一致していることをテーブル駆動で確認する。
// 3 サービス間でのヘッダ運用のドリフトをユニットテスト層で検出する。
func TestSecurityHeadersValues(t *testing.T) {
	expected := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "no-referrer",
	}
	if len(SecurityHeaders) != len(expected) {
		t.Fatalf("expected %d headers, got %d (drift from api-gateway / alert-service?)",
			len(expected), len(SecurityHeaders))
	}
	for name, want := range expected {
		got, ok := SecurityHeaders[name]
		if !ok {
			t.Errorf("missing security header %q", name)
			continue
		}
		if got != want {
			t.Errorf("header %q: expected %q, got %q", name, want, got)
		}
	}
}
