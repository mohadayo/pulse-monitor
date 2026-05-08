export type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';

const LEVEL_ORDER: Record<LogLevel, number> = {
  DEBUG: 0,
  INFO: 1,
  WARN: 2,
  ERROR: 3,
};

export class Logger {
  private level: LogLevel;
  private service: string;

  constructor(service: string, level: LogLevel = 'INFO') {
    this.service = service;
    this.level = level;
  }

  private shouldLog(level: LogLevel): boolean {
    return LEVEL_ORDER[level] >= LEVEL_ORDER[this.level];
  }

  private format(level: LogLevel, message: string, data?: Record<string, unknown>): string {
    const entry = {
      timestamp: new Date().toISOString(),
      level,
      service: this.service,
      message,
      ...data,
    };
    return JSON.stringify(entry);
  }

  debug(message: string, data?: Record<string, unknown>): void {
    if (this.shouldLog('DEBUG')) console.log(this.format('DEBUG', message, data));
  }

  info(message: string, data?: Record<string, unknown>): void {
    if (this.shouldLog('INFO')) console.log(this.format('INFO', message, data));
  }

  warn(message: string, data?: Record<string, unknown>): void {
    if (this.shouldLog('WARN')) console.warn(this.format('WARN', message, data));
  }

  error(message: string, data?: Record<string, unknown>): void {
    if (this.shouldLog('ERROR')) console.error(this.format('ERROR', message, data));
  }
}
