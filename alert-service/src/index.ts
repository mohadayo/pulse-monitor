import { app } from './app';
import { loadConfig } from './config';
import { Logger } from './logger';

const config = loadConfig();
const logger = new Logger('alert-service', config.logLevel as 'DEBUG' | 'INFO' | 'WARN' | 'ERROR');

app.listen(config.port, () => {
  logger.info(`Alert service started on port ${config.port}`);
});
