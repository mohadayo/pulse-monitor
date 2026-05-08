import { Logger } from './logger';

export interface Alert {
  id: string;
  serviceId: string;
  serviceName: string;
  status: 'triggered' | 'resolved';
  message: string;
  createdAt: string;
}

export interface AlertRule {
  id: string;
  serviceId: string;
  webhookUrl?: string;
  email?: string;
  createdAt: string;
}

export interface CreateAlertRuleInput {
  serviceId: string;
  webhookUrl?: string;
  email?: string;
}

let alertCounter = 0;
let ruleCounter = 0;

export class AlertStore {
  private alerts: Map<string, Alert> = new Map();
  private rules: Map<string, AlertRule> = new Map();
  private logger: Logger;

  constructor(logger: Logger) {
    this.logger = logger;
  }

  createRule(input: CreateAlertRuleInput): AlertRule {
    ruleCounter++;
    const rule: AlertRule = {
      id: `rule-${ruleCounter}`,
      serviceId: input.serviceId,
      webhookUrl: input.webhookUrl,
      email: input.email,
      createdAt: new Date().toISOString(),
    };
    this.rules.set(rule.id, rule);
    this.logger.info('Alert rule created', { ruleId: rule.id, serviceId: rule.serviceId });
    return rule;
  }

  getRules(): AlertRule[] {
    return Array.from(this.rules.values());
  }

  getRule(id: string): AlertRule | undefined {
    return this.rules.get(id);
  }

  deleteRule(id: string): boolean {
    const deleted = this.rules.delete(id);
    if (deleted) {
      this.logger.info('Alert rule deleted', { ruleId: id });
    }
    return deleted;
  }

  triggerAlert(serviceId: string, serviceName: string, message: string): Alert {
    alertCounter++;
    const alert: Alert = {
      id: `alert-${alertCounter}`,
      serviceId,
      serviceName,
      status: 'triggered',
      message,
      createdAt: new Date().toISOString(),
    };
    this.alerts.set(alert.id, alert);
    this.logger.warn('Alert triggered', { alertId: alert.id, serviceId, message });
    return alert;
  }

  getAlerts(): Alert[] {
    return Array.from(this.alerts.values());
  }

  resolveAlert(id: string): Alert | undefined {
    const alert = this.alerts.get(id);
    if (alert) {
      alert.status = 'resolved';
      this.logger.info('Alert resolved', { alertId: id });
    }
    return alert;
  }

  clear(): void {
    this.alerts.clear();
    this.rules.clear();
    alertCounter = 0;
    ruleCounter = 0;
  }
}
