import { AlertStore } from './alerts';
import { Logger } from './logger';

function makeStore(dedupWindowMs?: number): AlertStore {
  // ERROR level keeps test output clean while still exercising the logger path.
  const logger = new Logger('alert-service-test', 'ERROR');
  return new AlertStore(logger, dedupWindowMs === undefined ? {} : { dedupWindowMs });
}

describe('AlertStore rules', () => {
  it('creates, lists, retrieves, and deletes a rule', () => {
    const store = makeStore();
    store.clear();

    const rule = store.createRule({ serviceId: 'svc-1', webhookUrl: 'https://hooks.example.com' });
    expect(rule.id).toMatch(/^rule-/);
    expect(rule.serviceId).toBe('svc-1');
    expect(rule.createdAt).toBeDefined();

    expect(store.getRules()).toHaveLength(1);
    expect(store.getRule(rule.id)?.serviceId).toBe('svc-1');

    expect(store.deleteRule(rule.id)).toBe(true);
    expect(store.deleteRule(rule.id)).toBe(false);
    expect(store.getRules()).toHaveLength(0);
  });
});

describe('AlertStore.triggerAlert dedup', () => {
  it('deduplicates repeated alerts within the window and bumps count/updatedAt', async () => {
    const store = makeStore(60_000);
    store.clear();

    const first = store.triggerAlert('svc-1', 'Web', 'down');
    // Guarantee updatedAt clock ticks forward for the second call.
    await new Promise((resolve) => setTimeout(resolve, 5));
    const second = store.triggerAlert('svc-1', 'Web', 'down');

    expect(second.id).toBe(first.id);
    expect(second.count).toBe(2);
    expect(Date.parse(second.updatedAt)).toBeGreaterThanOrEqual(Date.parse(first.createdAt));
    expect(store.getAlerts()).toHaveLength(1);
  });

  it('creates a new alert when the message differs even if serviceId matches', () => {
    const store = makeStore(60_000);
    store.clear();

    const a = store.triggerAlert('svc-1', 'Web', 'down');
    const b = store.triggerAlert('svc-1', 'Web', 'high latency');

    expect(b.id).not.toBe(a.id);
    expect(store.getAlerts()).toHaveLength(2);
  });

  it('creates a new alert when the previous one is already resolved', () => {
    const store = makeStore(60_000);
    store.clear();

    const first = store.triggerAlert('svc-1', 'Web', 'down');
    store.resolveAlert(first.id);
    const second = store.triggerAlert('svc-1', 'Web', 'down');

    expect(second.id).not.toBe(first.id);
    expect(second.status).toBe('triggered');
    expect(second.count).toBe(1);
    expect(store.getAlerts()).toHaveLength(2);
  });

  it('creates a new alert when the previous fire is older than the dedup window', () => {
    // 1ms window forces the second fire to fall outside it after a small wait.
    const store = makeStore(1);
    store.clear();

    const first = store.triggerAlert('svc-1', 'Web', 'down');
    return new Promise<void>((resolve) => {
      setTimeout(() => {
        const second = store.triggerAlert('svc-1', 'Web', 'down');
        expect(second.id).not.toBe(first.id);
        expect(store.getAlerts()).toHaveLength(2);
        resolve();
      }, 20);
    });
  });

  it('disables dedup when the window is 0', () => {
    const store = makeStore(0);
    store.clear();

    const a = store.triggerAlert('svc-1', 'Web', 'down');
    const b = store.triggerAlert('svc-1', 'Web', 'down');
    const c = store.triggerAlert('svc-1', 'Web', 'down');

    expect(new Set([a.id, b.id, c.id]).size).toBe(3);
    expect(store.getAlerts()).toHaveLength(3);
  });
});

describe('AlertStore.resolveAlert', () => {
  it('marks an alert as resolved and refreshes updatedAt', async () => {
    const store = makeStore();
    store.clear();

    const alert = store.triggerAlert('svc-1', 'Web', 'down');
    const originalUpdatedAt = alert.updatedAt;
    await new Promise((resolve) => setTimeout(resolve, 5));

    const resolved = store.resolveAlert(alert.id);
    expect(resolved?.status).toBe('resolved');
    expect(Date.parse(resolved!.updatedAt)).toBeGreaterThanOrEqual(Date.parse(originalUpdatedAt));
  });

  it('returns undefined for an unknown id', () => {
    const store = makeStore();
    store.clear();
    expect(store.resolveAlert('does-not-exist')).toBeUndefined();
  });
});
