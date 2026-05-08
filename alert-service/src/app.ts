import express, { Request, Response, NextFunction } from 'express';
import { Logger } from './logger';
import { AlertStore, CreateAlertRuleInput } from './alerts';
import { loadConfig } from './config';

const config = loadConfig();
const logger = new Logger('alert-service', config.logLevel as 'DEBUG' | 'INFO' | 'WARN' | 'ERROR');
const store = new AlertStore(logger);

const app = express();
app.use(express.json());

app.get('/health', (_req: Request, res: Response) => {
  logger.info('Health check requested');
  res.json({
    status: 'ok',
    service: 'alert-service',
    version: '1.0.0',
    timestamp: new Date().toISOString(),
  });
});

app.post('/rules', (req: Request, res: Response) => {
  const input: CreateAlertRuleInput = req.body;
  if (!input.serviceId) {
    res.status(400).json({ error: 'serviceId is required' });
    return;
  }
  const rule = store.createRule(input);
  res.status(201).json(rule);
});

app.get('/rules', (_req: Request, res: Response) => {
  logger.info('Listing alert rules');
  res.json(store.getRules());
});

app.get('/rules/:id', (req: Request<{id: string}>, res: Response) => {
  const rule = store.getRule(req.params.id);
  if (!rule) {
    res.status(404).json({ error: 'Rule not found' });
    return;
  }
  res.json(rule);
});

app.delete('/rules/:id', (req: Request<{id: string}>, res: Response) => {
  if (!store.deleteRule(req.params.id)) {
    res.status(404).json({ error: 'Rule not found' });
    return;
  }
  res.status(204).send();
});

app.post('/alerts', (req: Request, res: Response) => {
  const { serviceId, serviceName, message } = req.body;
  if (!serviceId || !message) {
    res.status(400).json({ error: 'serviceId and message are required' });
    return;
  }
  const alert = store.triggerAlert(serviceId, serviceName || 'unknown', message);
  res.status(201).json(alert);
});

app.get('/alerts', (_req: Request, res: Response) => {
  logger.info('Listing alerts');
  res.json(store.getAlerts());
});

app.put('/alerts/:id/resolve', (req: Request<{id: string}>, res: Response) => {
  const alert = store.resolveAlert(req.params.id);
  if (!alert) {
    res.status(404).json({ error: 'Alert not found' });
    return;
  }
  res.json(alert);
});

app.use((err: Error, _req: Request, res: Response, _next: NextFunction) => {
  logger.error('Unhandled error', { error: err.message });
  res.status(500).json({ error: 'internal_server_error', detail: 'An unexpected error occurred' });
});

export { app, store };
