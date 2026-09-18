import { Request, Response, NextFunction } from 'express';

/**
 * `alert-service` の全応答に付与する最小限のセキュリティレスポンスヘッダ。
 *
 * `api-gateway` の `SecurityHeadersMiddleware` (`app/main.py`) と同一の
 * ポリシーを保ち、3 サービスでのヘッダ運用を揃える意図で設定している。
 *
 * - `X-Content-Type-Options: nosniff` … JSON エンドポイントを別 MIME として
 *   解釈させる MIME sniffing 攻撃を抑止する。
 * - `X-Frame-Options: DENY` … API を `<iframe>` に埋め込ませない。
 *   JSON API はフレーム表示を意図しないため常時拒否する (clickjacking 対策)。
 * - `Referrer-Policy: no-referrer` … 内部 URL やクエリ文字列が、リンク先の
 *   Referrer ヘッダとして外部に漏れないよう抑止する。
 *
 * 既にハンドラ側で同名ヘッダが設定されている場合は上書きしない
 * (`setdefault` 相当)。将来、特定ルートで独自のヘッダ値を返したくなった
 * 場合の逃げ道を残し、テストやミドルウェア追加時の後方互換を保つ。
 */
export const SECURITY_HEADERS: Readonly<Record<string, string>> = Object.freeze({
  'X-Content-Type-Options': 'nosniff',
  'X-Frame-Options': 'DENY',
  'Referrer-Policy': 'no-referrer',
});

export function securityHeaders() {
  return function securityHeadersMiddleware(
    _req: Request,
    res: Response,
    next: NextFunction,
  ): void {
    for (const [name, value] of Object.entries(SECURITY_HEADERS)) {
      // `getHeader` は大文字小文字を無視して既存ヘッダを返す。既に載って
      // いれば上書きしない (per-route オーバーライドを壊さない)。
      if (res.getHeader(name) === undefined) {
        res.setHeader(name, value);
      }
    }
    next();
  };
}
