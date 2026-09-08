import { loadConfig, resolveLogLevel, resolvePort } from './config';

const ENV_KEYS = ['ALERT_PORT', 'LOG_LEVEL', 'API_GATEWAY_URL', 'ALERT_DEDUP_WINDOW_SECONDS'];

/**
 * config は環境変数を都度読むため、各テストの前後で対象キーを退避・復元する。
 */
function withEnv<T>(overrides: Record<string, string | undefined>, fn: () => T): T {
  const saved: Record<string, string | undefined> = {};
  for (const key of ENV_KEYS) saved[key] = process.env[key];
  try {
    for (const [k, v] of Object.entries(overrides)) {
      if (v === undefined) delete process.env[k];
      else process.env[k] = v;
    }
    // Clear any keys not explicitly overridden so we exercise defaults.
    for (const key of ENV_KEYS) {
      if (!(key in overrides)) delete process.env[key];
    }
    return fn();
  } finally {
    for (const key of ENV_KEYS) {
      if (saved[key] === undefined) delete process.env[key];
      else process.env[key] = saved[key]!;
    }
  }
}

describe('resolvePort', () => {
  it.each([
    ['1', 1],
    ['8002', 8002],
    ['9001', 9001],
    ['65535', 65535],
  ])('accepts valid port %s', (raw, expected) => {
    expect(resolvePort(raw)).toBe(expected);
  });

  it.each([undefined, ''])('falls back to default when unset (%s)', (raw) => {
    expect(resolvePort(raw)).toBe(8002);
  });

  it.each(['abc', '8002abc', '8.5', '0x1F', '-'])(
    'falls back for non-integer %s and warns',
    (raw) => {
      const spy = jest.spyOn(process.stderr, 'write').mockImplementation(() => true);
      try {
        expect(resolvePort(raw)).toBe(8002);
        expect(spy).toHaveBeenCalled();
        const msg = spy.mock.calls.map((c) => String(c[0])).join('');
        expect(msg).toContain('invalid ALERT_PORT');
      } finally {
        spy.mockRestore();
      }
    },
  );

  it.each(['0', '-1', '65536', '999999'])(
    'falls back for out-of-range %s and warns',
    (raw) => {
      const spy = jest.spyOn(process.stderr, 'write').mockImplementation(() => true);
      try {
        expect(resolvePort(raw)).toBe(8002);
        expect(spy).toHaveBeenCalled();
        const msg = spy.mock.calls.map((c) => String(c[0])).join('');
        expect(msg).toContain('out of range');
      } finally {
        spy.mockRestore();
      }
    },
  );
});

describe('resolveLogLevel', () => {
  it.each([
    ['DEBUG', 'DEBUG'],
    ['INFO', 'INFO'],
    ['WARN', 'WARN'],
    ['ERROR', 'ERROR'],
    ['debug', 'DEBUG'],
    ['  info  ', 'INFO'],
    ['Error', 'ERROR'],
  ])('normalizes %s to %s', (raw, expected) => {
    expect(resolveLogLevel(raw)).toBe(expected);
  });

  it('aliases WARNING to WARN so ops can share LOG_LEVEL with api-gateway', () => {
    expect(resolveLogLevel('WARNING')).toBe('WARN');
    expect(resolveLogLevel('warning')).toBe('WARN');
  });

  it.each([undefined, ''])('falls back to INFO when unset (%s)', (raw) => {
    expect(resolveLogLevel(raw)).toBe('INFO');
  });

  it.each(['TRACE', 'verbose', 'SHOUT', '1', '!!'])(
    'falls back to INFO for invalid %s and warns',
    (raw) => {
      const spy = jest.spyOn(process.stderr, 'write').mockImplementation(() => true);
      try {
        expect(resolveLogLevel(raw)).toBe('INFO');
        expect(spy).toHaveBeenCalled();
        const msg = spy.mock.calls.map((c) => String(c[0])).join('');
        expect(msg).toContain('invalid LOG_LEVEL');
      } finally {
        spy.mockRestore();
      }
    },
  );
});

describe('loadConfig', () => {
  it('returns defaults when env vars are absent', () => {
    withEnv({}, () => {
      const cfg = loadConfig();
      expect(cfg.port).toBe(8002);
      expect(cfg.logLevel).toBe('INFO');
      expect(cfg.apiGatewayUrl).toBe('http://api-gateway:8000');
      expect(cfg.alertDedupWindowSeconds).toBe(300);
    });
  });

  it('accepts valid overrides', () => {
    withEnv(
      {
        ALERT_PORT: '9002',
        LOG_LEVEL: 'debug',
        API_GATEWAY_URL: 'http://gw:1',
        ALERT_DEDUP_WINDOW_SECONDS: '60',
      },
      () => {
        const cfg = loadConfig();
        expect(cfg.port).toBe(9002);
        expect(cfg.logLevel).toBe('DEBUG');
        expect(cfg.apiGatewayUrl).toBe('http://gw:1');
        expect(cfg.alertDedupWindowSeconds).toBe(60);
      },
    );
  });

  it('does not throw and yields usable defaults when env is malformed', () => {
    // Regression guard: previously ALERT_PORT=abc silently produced NaN and
    // LOG_LEVEL=verbose silently disabled the logger.
    const spy = jest.spyOn(process.stderr, 'write').mockImplementation(() => true);
    try {
      withEnv({ ALERT_PORT: 'not-a-port', LOG_LEVEL: 'verbose' }, () => {
        const cfg = loadConfig();
        expect(cfg.port).toBe(8002);
        expect(cfg.logLevel).toBe('INFO');
      });
    } finally {
      spy.mockRestore();
    }
  });
});
