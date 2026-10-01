package server

import "net/http"

// SecurityHeaders は health-checker の全応答に付与する最小限の
// セキュリティレスポンスヘッダ。
//
// `api-gateway` の `SecurityHeadersMiddleware` (`app/main.py`) および
// `alert-service` の `SECURITY_HEADERS` (`src/security-headers.ts`) と
// 同一のポリシーを保ち、3 サービスでのヘッダ運用を揃える意図で設定している。
//
// - `X-Content-Type-Options: nosniff` … JSON エンドポイントを別 MIME として
//   解釈させる MIME sniffing 攻撃を抑止する。
// - `X-Frame-Options: DENY` … API を `<iframe>` に埋め込ませない。
//   JSON API はフレーム表示を意図しないため常時拒否する (clickjacking 対策)。
// - `Referrer-Policy: no-referrer` … 内部 URL やクエリ文字列が、リンク先の
//   Referrer ヘッダとして外部に漏れないよう抑止する。
var SecurityHeaders = map[string]string{
	"X-Content-Type-Options": "nosniff",
	"X-Frame-Options":        "DENY",
	"Referrer-Policy":        "no-referrer",
}

// securityHeadersMiddleware は SecurityHeaders を全応答に追加する chi 互換ミドルウェア。
//
// 既にハンドラが同名ヘッダを設定している場合は上書きしない (`setdefault` 相当)。
// 将来、特定ルートで独自の値を返したくなった場合の逃げ道を残し、既存テストや
// ミドルウェア追加時の後方互換を保つ。
//
// chi.Mux では `router.Use(...)` で登録することで、ルーティング直前のパイプラインに
// 挿入され、ハンドラ側の `w.WriteHeader(...)` より前にヘッダ辞書へ書き込める。
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		for name, value := range SecurityHeaders {
			if h.Get(name) == "" {
				h.Set(name, value)
			}
		}
		next.ServeHTTP(w, r)
	})
}
