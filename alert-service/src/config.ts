export interface Config {
  port: number;
  logLevel: string;
  apiGatewayUrl: string;
  alertDedupWindowSeconds: number;
}

const DEFAULT_DEDUP_WINDOW_SECONDS = 300;

function parseNonNegativeInt(raw: string | undefined, fallback: number): number {
  if (raw === undefined || raw === '') return fallback;
  const parsed = parseInt(raw, 10);
  if (Number.isNaN(parsed) || parsed < 0) return fallback;
  return parsed;
}

export function loadConfig(): Config {
  return {
    port: parseInt(process.env.ALERT_PORT || '8002', 10),
    logLevel: process.env.LOG_LEVEL || 'INFO',
    apiGatewayUrl: process.env.API_GATEWAY_URL || 'http://api-gateway:8000',
    alertDedupWindowSeconds: parseNonNegativeInt(
      process.env.ALERT_DEDUP_WINDOW_SECONDS,
      DEFAULT_DEDUP_WINDOW_SECONDS,
    ),
  };
}
