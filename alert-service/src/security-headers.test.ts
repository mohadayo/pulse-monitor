import express from 'express';
import request from 'supertest';
import { app, store } from './app';
import { SECURITY_HEADERS, securityHeaders } from './security-headers';

beforeEach(() => {
  store.clear();
});

/**
 * `securityHeaders()` ミドルウェアの回帰テスト。
 *
 * `alert-service` の全応答に付与される最小限のセキュリティヘッダを
 * 成功系 (200 / 201) / 未定義パス (404) / エラーレスポンス (400) の
 * 3 経路で固定し、「特定ルートだけ抜ける」リグレッションを検出する。
 * api-gateway の `test_security_headers.py` と同じ意図。
 */
function assertSecurityHeaders(res: request.Response): void {
  expect(res.headers['x-content-type-options']).toBe('nosniff');
  expect(res.headers['x-frame-options']).toBe('DENY');
  expect(res.headers['referrer-policy']).toBe('no-referrer');
}

describe('securityHeaders middleware (on app)', () => {
  it('sets headers on 200 responses (/health)', async () => {
    const res = await request(app).get('/health');
    expect(res.status).toBe(200);
    assertSecurityHeaders(res);
  });

  it('sets headers on 201 responses (POST /rules)', async () => {
    const res = await request(app)
      .post('/rules')
      .send({ serviceId: 'svc-1', webhookUrl: 'https://hooks.example.com' });
    expect(res.status).toBe(201);
    assertSecurityHeaders(res);
  });

  it('sets headers on 404 responses for undefined paths', async () => {
    const res = await request(app).get('/definitely-not-a-real-endpoint');
    expect(res.status).toBe(404);
    assertSecurityHeaders(res);
  });

  it('sets headers on 404 for missing resource (unknown rule id)', async () => {
    const res = await request(app).get('/rules/does-not-exist');
    expect(res.status).toBe(404);
    assertSecurityHeaders(res);
  });

  it('sets headers on 400 validation errors', async () => {
    const res = await request(app).post('/rules').send({});
    expect(res.status).toBe(400);
    assertSecurityHeaders(res);
  });
});

describe('securityHeaders middleware (unit, does not override existing values)', () => {
  it('leaves per-route custom values intact', async () => {
    // 独立した小さな Express アプリで、ルート側で先に同名ヘッダを設定した
    // 場合にミドルウェアが上書きしないこと (setdefault 相当) を検証する。
    // 将来、特定ルートで独自のヘッダ値を返したくなった場合の逃げ道を保証。
    const testApp = express();
    testApp.use(securityHeaders());
    testApp.get('/custom', (_req, res) => {
      res.setHeader('X-Content-Type-Options', 'custom-value');
      res.status(200).json({ ok: true });
    });

    const res = await request(testApp).get('/custom');
    expect(res.status).toBe(200);
    // ルート側の "custom-value" が保持される
    expect(res.headers['x-content-type-options']).toBe('custom-value');
    // 他 2 ヘッダはミドルウェアが付与
    expect(res.headers['x-frame-options']).toBe('DENY');
    expect(res.headers['referrer-policy']).toBe('no-referrer');
  });
});

describe('SECURITY_HEADERS constants', () => {
  it('is frozen so downstream code cannot mutate the shared policy', () => {
    expect(Object.isFrozen(SECURITY_HEADERS)).toBe(true);
  });

  it('exposes exactly the three JSON-API-safe headers', () => {
    expect(SECURITY_HEADERS).toEqual({
      'X-Content-Type-Options': 'nosniff',
      'X-Frame-Options': 'DENY',
      'Referrer-Policy': 'no-referrer',
    });
  });
});
