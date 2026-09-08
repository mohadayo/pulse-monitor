import { LogLevel } from './logger';

export interface Config {
  port: number;
  logLevel: LogLevel;
  apiGatewayUrl: string;
  alertDedupWindowSeconds: number;
}

const DEFAULT_PORT = 8002;
const MIN_PORT = 1;
const MAX_PORT = 65535;
const DEFAULT_LOG_LEVEL: LogLevel = 'INFO';
const ALLOWED_LOG_LEVELS: readonly LogLevel[] = ['DEBUG', 'INFO', 'WARN', 'ERROR'];
// `WARNING` は Python 側 (api-gateway) の別名。運用者が揃った命名で
// 指定できるよう受理し、正式名 `WARN` に寄せて logger に渡す。
const LOG_LEVEL_ALIASES: Record<string, LogLevel> = { WARNING: 'WARN' };
const DEFAULT_DEDUP_WINDOW_SECONDS = 300;

/**
 * ALERT_PORT を 1〜65535 の整数として解決する。
 *
 * 数値化に失敗した場合や範囲外の場合はデフォルト (8002) を返し、
 * 運用者が原因を追えるよう stderr に警告を書き出す。プロセスは落とさない。
 * これまでは ``parseInt(process.env.ALERT_PORT || '8002', 10)`` のみで
 * ``ALERT_PORT=abc`` のような値は ``NaN`` になり、express の
 * ``listen(NaN)`` が実質ランダムポート (0) を掴む挙動になっていた。
 */
export function resolvePort(raw: string | undefined): number {
  if (raw === undefined || raw === '') return DEFAULT_PORT;
  // parseInt は末尾のゴミ (`"8002abc"`) を許容してしまうため、厳密に判定する。
  if (!/^-?\d+$/.test(raw.trim())) {
    process.stderr.write(
      `[config] WARNING: invalid ALERT_PORT=${JSON.stringify(raw)}; ` +
        `falling back to ${DEFAULT_PORT}.\n`,
    );
    return DEFAULT_PORT;
  }
  const parsed = parseInt(raw, 10);
  if (Number.isNaN(parsed) || parsed < MIN_PORT || parsed > MAX_PORT) {
    process.stderr.write(
      `[config] WARNING: ALERT_PORT=${parsed} is out of range ` +
        `[${MIN_PORT}, ${MAX_PORT}]; falling back to ${DEFAULT_PORT}.\n`,
    );
    return DEFAULT_PORT;
  }
  return parsed;
}

/**
 * LOG_LEVEL を Logger が理解する 4 値 (DEBUG/INFO/WARN/ERROR) に正規化する。
 *
 * 未知の値を渡すとデフォルト (INFO) にフォールバックし、stderr に警告を出す。
 * これまでは ``LOG_LEVEL || 'INFO'`` を素通しで Logger に渡し、
 * ``Logger.shouldLog`` 内の ``LEVEL_ORDER[this.level]`` が ``undefined`` を
 * 返した結果、``>=`` 比較が常に ``false`` となり **全てのログが黙って
 * 落ちる** 状態になっていた。
 */
export function resolveLogLevel(raw: string | undefined): LogLevel {
  if (raw === undefined || raw === '') return DEFAULT_LOG_LEVEL;
  const normalized = raw.trim().toUpperCase();
  if (normalized in LOG_LEVEL_ALIASES) return LOG_LEVEL_ALIASES[normalized];
  if ((ALLOWED_LOG_LEVELS as readonly string[]).includes(normalized)) {
    return normalized as LogLevel;
  }
  process.stderr.write(
    `[config] WARNING: invalid LOG_LEVEL=${JSON.stringify(raw)}; ` +
      `falling back to ${DEFAULT_LOG_LEVEL}. ` +
      `Allowed values: ${JSON.stringify([...ALLOWED_LOG_LEVELS])}.\n`,
  );
  return DEFAULT_LOG_LEVEL;
}

function parseNonNegativeInt(raw: string | undefined, fallback: number): number {
  if (raw === undefined || raw === '') return fallback;
  const parsed = parseInt(raw, 10);
  if (Number.isNaN(parsed) || parsed < 0) return fallback;
  return parsed;
}

export function loadConfig(): Config {
  return {
    port: resolvePort(process.env.ALERT_PORT),
    logLevel: resolveLogLevel(process.env.LOG_LEVEL),
    apiGatewayUrl: process.env.API_GATEWAY_URL || 'http://api-gateway:8000',
    alertDedupWindowSeconds: parseNonNegativeInt(
      process.env.ALERT_DEDUP_WINDOW_SECONDS,
      DEFAULT_DEDUP_WINDOW_SECONDS,
    ),
  };
}
