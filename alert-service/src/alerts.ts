import { Logger } from './logger';

export interface Alert {
  id: string;
  serviceId: string;
  serviceName: string;
  status: 'triggered' | 'resolved';
  message: string;
  count: number;
  createdAt: string;
  updatedAt: string;
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

export interface AlertStoreOptions {
  /**
   * Dedup / backoff window in milliseconds. When a duplicate alert
   * (same serviceId + message that is still `triggered`) arrives within
   * this window, the existing record's `count` and `updatedAt` are
   * bumped instead of a new alert being created. Set to `0` to disable.
   */
  dedupWindowMs?: number;
}

const DEFAULT_DEDUP_WINDOW_MS = 5 * 60 * 1000;

let alertCounter = 0;
let ruleCounter = 0;

export class AlertStore {
  private alerts: Map<string, Alert> = new Map();
  private rules: Map<string, AlertRule> = new Map();
  private logger: Logger;
  private dedupWindowMs: number;

  constructor(logger: Logger, options: AlertStoreOptions = {}) {
    this.logger = logger;
    const window = options.dedupWindowMs ?? DEFAULT_DEDUP_WINDOW_MS;
    // Reject negatives; treat as disabled.
    this.dedupWindowMs = window < 0 ? 0 : window;
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
    if (this.dedupWindowMs > 0) {
      const existing = this.findDedupCandidate(serviceId, message);
      if (existing) {
        existing.count += 1;
        existing.updatedAt = new Date().toISOString();
        this.logger.info('Alert deduplicated', {
          alertId: existing.id,
          serviceId,
          count: existing.count,
        });
        return existing;
      }
    }

    alertCounter++;
    const now = new Date().toISOString();
    const alert: Alert = {
      id: `alert-${alertCounter}`,
      serviceId,
      serviceName,
      status: 'triggered',
      message,
      count: 1,
      createdAt: now,
      updatedAt: now,
    };
    this.alerts.set(alert.id, alert);
    this.logger.warn('Alert triggered', { alertId: alert.id, serviceId, message });
    return alert;
  }

  private findDedupCandidate(serviceId: string, message: string): Alert | undefined {
    const cutoff = Date.now() - this.dedupWindowMs;
    for (const alert of this.alerts.values()) {
      if (
        alert.status === 'triggered' &&
        alert.serviceId === serviceId &&
        alert.message === message &&
        Date.parse(alert.updatedAt) >= cutoff
      ) {
        return alert;
      }
    }
    return undefined;
  }

  getAlerts(): Alert[] {
    return Array.from(this.alerts.values());
  }

  resolveAlert(id: string): Alert | undefined {
    const alert = this.alerts.get(id);
    if (alert) {
      alert.status = 'resolved';
      alert.updatedAt = new Date().toISOString();
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
