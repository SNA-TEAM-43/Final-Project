# RabbitMQ (notification buffer between services and Telegram)

**RabbitMQ** is the **async message broker** between **producers** (the Go API for application-level critical events; optionally a thin **Alertmanager webhook listener** that enqueues Prometheus alerts) and the **consumer** (**Telegram notifier worker** Bot API calls).

This decouples slow or flaky downstream delivery from request paths: HTTP handlers acknowledge quickly while **persisted queues** soak spikes; Telegram rate limits **back-pressure** onto the worker’s prefetch, not onto the API.

---

## Roles in the pipeline

| Component | Direction | Responsibility |
| --- | --- | --- |
| **Go HTTP backend** | **Producer** | Publishes small JSON payloads (severity, summary, correlation id) for **domain events** worth paging—for example repeated upload failures—without calling Telegram synchronously. |
| **Webhook-to-AMQP service** *(optional)* | **Producer** | Receives **`POST`** from Alertmanager, validates source, publishes normalized messages to the same topology as the API (**one queue for Telegram worker** simplifies operations). |
| **RabbitMQ** | Broker | **Durable queues** (and optional DLQ for poison messages); **TLS** optional on LAN-only Pi. |
| **Telegram notifier worker** *(Go binary)* | **Consumer** | Subscribes (competing consumers if scaled); transforms JSON → Telegram `sendMessage`; applies **retry** / **publisher confirms** semantics as you implement. |

Only the **worker** holds `TELEGRAM_BOT_TOKEN` if you want smallest blast radius—the API then never sees the bot secret.

---

## Suggested topology (exchange and queue)

Starter pattern (adjust names in implementation):

| Construct | Declaration | Typical binding |
| --- | --- | --- |
| Exchange | `notifications` — type **topic** or **direct** | — |
| Queue | `telegram.outbound.durable` | `durable: true`; consumer fair dispatch `prefetch=4`–`16` |
| Routing keys | `event.critical`, `alert.fired`, `alert.resolved` | Bind queue to routing keys your producers use |

Alternatively use a **fanout** exchange if every Telegram worker receives all alerts (usually one worker is enough on a Pi).

**Persistence**

- Declare **durable queues** and publish with **persistent** delivery mode (`delivery_mode = 2` in AMQP 0.9.1 terminology) when loss on broker restart during an outage matters.
- On **Raspberry Pi + SD**, place RabbitMQ **`data`** on **USB SSD** if you enqueue heavily in labs.

---

## Environment variables

| Variable | Used by | Role |
| --- | --- | --- |
| `RABBITMQ_URL` | API, webhook publisher, notifier worker | **`amqp://user:pass@rabbitmq:5672/`** in Compose (`rabbitmq` = service hostname). |
| `NOTIFY_EXCHANGE` | Producers | Exchange name (`notifications`). |
| `NOTIFY_ROUTING_KEY` | Producers | Default key if not set per-event (`event.critical`). |
| `TELEGRAM_BOT_TOKEN` | Worker only | Telegram BotFather token (**do not mount into API** if API only publishes to AMQP). |
| `TELEGRAM_CHAT_ID` | Worker only | Recipient chat or channel id. |

**Readiness**

- Optionally include **RabbitMQ reachability** in API **`/ready`** when publishing is mandatory for correctness; alternatively degrade to **metrics + logs only** when broker is down and document the trade-off.

See [Telegram notifications](./telegram.md) for receiver behaviour after dequeue.

---

## Failure and demo narratives

| Fault | Behaviour to highlight |
| --- | --- |
| Worker offline | Messages **accumulate** in RabbitMQ; recovery drains backlog—visual in management UI (**15672**) or queue depth metric. |
| Telegram API outage | Worker retries/backoff; queue grows; correlate with Grafana **publisher lag** metric if instrumented. |
| Pi network down | Producer cannot connect—**timeouts** logged; Prometheus still scrapes `/metrics`; alerts may still enqueue when network returns depending onpublisher design. |

---

## Docker Compose hints

Official image **`rabbitmq:3-management-alpine`** (arm64-capable):

- Ports **5672** (AMQP) internal on Docker network—**expose** `:15672` to LAN only when you demo the UI, or **`127.0.0.1:15672` + SSH tunnel**.
- Persist `/var/lib/rabbitmq`.

Add services **`rabbitmq`**, **`telegram-notifier`** (build from `cmd/telegram-notifier` or similar), **`alert-amqp-gateway`** *(optional)*.

Cross-links: **[Edge lab host](./edge-lab-host.md)** deployment, **[Go backend](./go-backend.md)** producer hooks, **[Prometheus](./prometheus.md)** + Alertmanager webhook path.
