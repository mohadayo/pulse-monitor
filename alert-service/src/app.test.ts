import request from 'supertest';
import { app, store } from './app';

beforeEach(() => {
  store.clear();
});

describe('GET /health', () => {
  it('returns healthy status', async () => {
    const res = await request(app).get('/health');
    expect(res.status).toBe(200);
    expect(res.body.status).toBe('ok');
    expect(res.body.service).toBe('alert-service');
    expect(res.body.version).toBe('1.0.0');
    expect(res.body.timestamp).toBeDefined();
  });
});

describe('POST /rules', () => {
  it('creates an alert rule', async () => {
    const res = await request(app)
      .post('/rules')
      .send({ serviceId: 'svc-1', webhookUrl: 'https://hooks.example.com/alert' });
    expect(res.status).toBe(201);
    expect(res.body.serviceId).toBe('svc-1');
    expect(res.body.webhookUrl).toBe('https://hooks.example.com/alert');
    expect(res.body.id).toBeDefined();
  });

  it('rejects missing serviceId', async () => {
    const res = await request(app).post('/rules').send({});
    expect(res.status).toBe(400);
  });
});

describe('GET /rules', () => {
  it('lists all rules', async () => {
    await request(app).post('/rules').send({ serviceId: 'svc-1' });
    await request(app).post('/rules').send({ serviceId: 'svc-2' });
    const res = await request(app).get('/rules');
    expect(res.status).toBe(200);
    expect(res.body).toHaveLength(2);
  });
});

describe('GET /rules/:id', () => {
  it('returns a rule by id', async () => {
    const created = await request(app).post('/rules').send({ serviceId: 'svc-1' });
    const res = await request(app).get(`/rules/${created.body.id}`);
    expect(res.status).toBe(200);
    expect(res.body.serviceId).toBe('svc-1');
  });

  it('returns 404 for unknown rule', async () => {
    const res = await request(app).get('/rules/unknown');
    expect(res.status).toBe(404);
  });
});

describe('DELETE /rules/:id', () => {
  it('deletes a rule', async () => {
    const created = await request(app).post('/rules').send({ serviceId: 'svc-1' });
    const res = await request(app).delete(`/rules/${created.body.id}`);
    expect(res.status).toBe(204);
  });

  it('returns 404 for unknown rule', async () => {
    const res = await request(app).delete('/rules/unknown');
    expect(res.status).toBe(404);
  });
});

describe('POST /alerts', () => {
  it('triggers an alert', async () => {
    const res = await request(app)
      .post('/alerts')
      .send({ serviceId: 'svc-1', serviceName: 'Web App', message: 'Service is down' });
    expect(res.status).toBe(201);
    expect(res.body.status).toBe('triggered');
    expect(res.body.message).toBe('Service is down');
  });

  it('rejects missing fields', async () => {
    const res = await request(app).post('/alerts').send({});
    expect(res.status).toBe(400);
  });
});

describe('GET /alerts', () => {
  it('lists all alerts', async () => {
    await request(app).post('/alerts').send({ serviceId: 'svc-1', message: 'down' });
    const res = await request(app).get('/alerts');
    expect(res.status).toBe(200);
    expect(res.body).toHaveLength(1);
  });
});

describe('PUT /alerts/:id/resolve', () => {
  it('resolves an alert', async () => {
    const created = await request(app).post('/alerts').send({ serviceId: 'svc-1', message: 'down' });
    const res = await request(app).put(`/alerts/${created.body.id}/resolve`);
    expect(res.status).toBe(200);
    expect(res.body.status).toBe('resolved');
  });

  it('returns 404 for unknown alert', async () => {
    const res = await request(app).put('/alerts/unknown/resolve');
    expect(res.status).toBe(404);
  });
});
