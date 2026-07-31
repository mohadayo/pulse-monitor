import { app } from './app';
import { loadConfig } from './config';
import { Logger } from './logger';

const config = loadConfig();
const logger = new Logger('alert-service', config.logLevel as 'DEBUG' | 'INFO' | 'WARN' | 'ERROR');

// SIGTERM 受信後、進行中リクエストの完了を待つ最大時間 (ms)。
// Kubernetes の既定 grace period (30s) より短めに設定し、SIGKILL が
// 飛ぶ前に確実にプロセスを終了させる。
function parseShutdownTimeoutMs(): number {
  const raw = process.env.SHUTDOWN_TIMEOUT_MS;
  if (raw === undefined || raw === '') {
    return 10000;
  }
  const parsed = parseInt(raw, 10);
  if (!Number.isFinite(parsed) || parsed <= 0) {
    logger.warn(`Invalid SHUTDOWN_TIMEOUT_MS=${raw}, falling back to 10000`);
    return 10000;
  }
  return parsed;
}

const server = app.listen(config.port, () => {
  logger.info(`Alert service started on port ${config.port}`);
});

const shutdownTimeoutMs = parseShutdownTimeoutMs();
let shuttingDown = false;
// SIGTERM / SIGINT を受けたら新規接続の受付を止め、進行中リクエストを
// 完了させたうえで終了する。タイムアウト超過時は強制終了 (exit 1)。
// Docker/Kubernetes 環境でのローリングアップデート中に、リクエストの
// 途中切断や 5xx を発生させないための実装。
function shutdown(signal: string): void {
  if (shuttingDown) {
    // 二重シグナル（SIGTERM→SIGINT 等）を受けても close コールバックが
    // 1 回しか呼ばれないよう、二重進入を弾く。
    return;
  }
  shuttingDown = true;
  logger.info(`Received ${signal}, shutting down gracefully`);
  const forceExit = setTimeout(() => {
    logger.error(`Graceful shutdown timed out after ${shutdownTimeoutMs}ms, forcing exit`);
    process.exit(1);
  }, shutdownTimeoutMs);
  // 待機用タイマー自体でイベントループを塞がないよう unref する。
  forceExit.unref();

  server.close((err) => {
    clearTimeout(forceExit);
    if (err) {
      logger.error('Error during server shutdown', { error: err.message });
      process.exit(1);
    }
    logger.info('Server closed cleanly');
    process.exit(0);
  });
}

process.on('SIGTERM', () => shutdown('SIGTERM'));
process.on('SIGINT', () => shutdown('SIGINT'));
