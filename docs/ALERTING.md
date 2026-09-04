# Alerting Guide

This document describes how alerting works in Pulse Monitor: which endpoints
the `alert-service` exposes, how an alert flows from detection to resolution,
how the deduplication window suppresses noisy repeats, and how to model
severity and escalation for on-call teams.

For a high-level architecture overview, see
[`ARCHITECTURE.md`](./ARCHITECTURE.md). For symptom-based debugging, see
[`TROUBLESHOOTING.md`](./TROUBLESHOOTING.md).

## Overview

`alert-service` is the TypeScript/Express component responsible for two
responsibilities:

1. **Alert rules** — declarative bindings between a `serviceId` and one or
   more notification channels (webhook, email).
2. **Alerts** — events raised (typically by `health-checker`) when a service
   is unhealthy, and closed when it recovers.

All state currently lives in-memory (see `alert-service/src/alerts.ts`); it is
rebuilt when the service restarts. Persistent storage is out of scope for
this document.

### Component interaction

```
              +-----------------+
              |   api-gateway   |
              |   (FastAPI)     |
              +--------+--------+
                       |
         proxies /alerts, /rules
                       v
+----------------+   POST /alerts   +------------------+
| health-checker | ---------------> |  alert-service   |
|      (Go)      |                  |  (TypeScript)    |
+----------------+                  +------------------+
                                            |
                                            | notify (future)
                                            v
                                    webhook / email
```

## HTTP endpoints

All endpoints are served by `alert-service` (default port `8002`) and are
also reachable through `api-gateway` on port `8000`. Request and response
bodies are JSON.

### Alert rules

| Method   | Path              | Purpose                                  |
| -------- | ----------------- | ---------------------------------------- |
| `POST`   | `/rules`          | Create an alert rule                     |
| `GET`    | `/rules`          | List all alert rules                     |
| `GET`    | `/rules/:id`      | Fetch a single rule                      |
| `DELETE` | `/rules/:id`      | Delete a rule (returns `204` on success) |

**`POST /rules` request body**

```json
{
  "serviceId": "api-gateway",
  "webhookUrl": "https://hooks.example.com/pulse",
  "email": "oncall@example.com"
}
```

- `serviceId` (required) — identifier of the service the rule watches.
- `webhookUrl` (optional) — destination for webhook notifications.
- `email` (optional) — destination for email notifications.

A `400` is returned when `serviceId` is missing. At least one of
`webhookUrl` or `email` should be provided for the rule to be actionable.

**Response `201`**

```json
{
  "id": "rule-1",
  "serviceId": "api-gateway",
  "webhookUrl": "https://hooks.example.com/pulse",
  "email": "oncall@example.com",
  "createdAt": "2026-09-04T09:00:00.000Z"
}
```

### Alerts

| Method | Path                    | Purpose                                        |
| ------ | ----------------------- | ---------------------------------------------- |
| `POST` | `/alerts`               | Trigger a new alert (or bump a deduped one)    |
| `GET`  | `/alerts`               | List all alerts (`triggered` and `resolved`)   |
| `PUT`  | `/alerts/:id/resolve`   | Mark an alert as `resolved`                    |

**`POST /alerts` request body**

```json
{
  "serviceId": "api-gateway",
  "serviceName": "API Gateway",
  "message": "health check failed: connection refused"
}
```

- `serviceId` (required)
- `message` (required) — human-readable reason; also acts as the dedup key.
- `serviceName` (optional) — defaults to `"unknown"`.

A `400` is returned when `serviceId` or `message` is missing.

**Response `201`**

```json
{
  "id": "alert-42",
  "serviceId": "api-gateway",
  "serviceName": "API Gateway",
  "status": "triggered",
  "message": "health check failed: connection refused",
  "count": 1,
  "createdAt": "2026-09-04T09:00:00.000Z",
  "updatedAt": "2026-09-04T09:00:00.000Z"
}
```

When the alert matches an active one within the dedup window, the existing
record is returned with an incremented `count` and refreshed `updatedAt`;
no new record is created (see [Deduplication](#deduplication)).

## Lifecycle

An alert moves through two states:

```
         POST /alerts               PUT /alerts/:id/resolve
(new) ------------------> triggered ------------------------> resolved
                            ^   |
     duplicate within window |   | (count++, updatedAt refreshed)
                            +---+
```

- **`triggered`** — initial state when a fresh alert is created.
- **`resolved`** — terminal state set by `PUT /alerts/:id/resolve`. Once
  resolved, an alert no longer participates in deduplication; a new incoming
  event with the same `serviceId` and `message` produces a fresh alert.

Every transition updates `updatedAt`. `count` reflects the total number of
raise attempts (initial + deduped duplicates) recorded against the alert.

## Deduplication

To suppress alert storms when a service flaps or when `health-checker`
retries an unhealthy target, `alert-service` deduplicates repeats within a
configurable window.

A new `POST /alerts` is considered a duplicate when **all** of the following
hold against an existing alert:

- `status` is still `triggered` (not yet resolved).
- `serviceId` matches exactly.
- `message` matches exactly (case-sensitive).
- The existing alert's `updatedAt` is within the window relative to now.

When matched, the existing alert has its `count` incremented and its
`updatedAt` refreshed; no new alert record is stored.

### Configuration

The window is controlled by the `ALERT_DEDUP_WINDOW_SECONDS` environment
variable (see [`.env.example`](../.env.example)).

| Value        | Meaning                                          |
| ------------ | ------------------------------------------------ |
| `300` (default) | Suppress duplicates for 5 minutes             |
| `0`          | Disable deduplication                            |
| `> 0`        | Any positive number of seconds                   |

Negative values are coerced to `0` (dedup disabled) at load time.

### When to tune it

- **Increase** (e.g. `900`–`3600`) if a single flapping dependency floods
  on-call with the same message; a longer window reduces pager fatigue.
- **Decrease** (e.g. `30`–`60`) when the alert message is coarse and you
  want faster re-notification after transient recovery.
- **Disable** (`0`) for smoke tests and integration environments where you
  want every event captured verbatim.

## Severity and escalation

Severity is not carried as a field on the alert payload today; the
recommended convention is to encode it into `message` and to route with
separate rules per `serviceId`. Suggested tiers:

| Severity | Meaning                                       | Response target |
| -------- | --------------------------------------------- | --------------- |
| **P1**   | User-visible outage or data loss risk         | Page immediately, 24/7 |
| **P2**   | Degraded functionality; SLO burn accelerating | Page during business hours, ticket after-hours |
| **P3**   | Non-user-visible warning or capacity signal   | Ticket only |

Recommended escalation policy (implemented by the receiving webhook or a
downstream notifier, not by `alert-service` itself):

1. **T+0** — notify primary on-call via webhook.
2. **T+15 min** — if `count` continues to grow and no `resolved` transition
   has occurred, notify secondary on-call.
3. **T+30 min** — notify the service owner group (email channel).
4. **T+60 min** — raise to incident channel and open an incident record.

## Notification channels

Rules can carry either a `webhookUrl`, an `email`, or both. Downstream
notification delivery is expected to be handled outside `alert-service`; the
service's responsibility is to persist rules and expose alert state.

**Register a webhook rule**

```bash
curl -X POST http://localhost:8002/rules \
  -H 'Content-Type: application/json' \
  -d '{
    "serviceId": "api-gateway",
    "webhookUrl": "https://hooks.example.com/pulse"
  }'
```

**Register an email rule**

```bash
curl -X POST http://localhost:8002/rules \
  -H 'Content-Type: application/json' \
  -d '{
    "serviceId": "api-gateway",
    "email": "oncall@example.com"
  }'
```

## Quickstart

With the stack running (`docker compose up`), the following sequence
exercises the full lifecycle end-to-end.

```bash
# 1. Register a webhook rule
curl -sS -X POST http://localhost:8002/rules \
  -H 'Content-Type: application/json' \
  -d '{"serviceId":"api-gateway","webhookUrl":"https://hooks.example.com/pulse"}'

# 2. Trigger an alert
curl -sS -X POST http://localhost:8002/alerts \
  -H 'Content-Type: application/json' \
  -d '{"serviceId":"api-gateway","serviceName":"API Gateway","message":"health check failed"}'

# 3. Trigger the same alert again — count increments, no new record
curl -sS -X POST http://localhost:8002/alerts \
  -H 'Content-Type: application/json' \
  -d '{"serviceId":"api-gateway","serviceName":"API Gateway","message":"health check failed"}'

# 4. Inspect current alerts
curl -sS http://localhost:8002/alerts | jq

# 5. Resolve it
curl -sS -X PUT http://localhost:8002/alerts/alert-1/resolve
```

## See also

- [`ARCHITECTURE.md`](./ARCHITECTURE.md) — service topology and data flow
- [`TROUBLESHOOTING.md`](./TROUBLESHOOTING.md) — symptom-based debugging
- [`FAQ.md`](./FAQ.md) — common questions
- [`../.env.example`](../.env.example) — all environment variables
