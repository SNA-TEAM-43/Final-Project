# Telegram notifications (via RabbitMQ worker)

Human-readable **critical** signals land in Telegram through a **`sendMessage`** call. In this architecture the **Telegram Bot API is called only by a dedicated notifier worker**, not synchronously inside the HTTP request path.

**Recommended path:**

```
Producer (Go API and/or webhook) → RabbitMQ → Go Telegram notifier worker → Telegram Bot API (cloud)
```

This matches the goal of buffering, **isolating Telegram rate limits** from latency-sensitive handlers, and **avoiding coupling** Prometheus/Alertmanager text formats directly to Telegram formatting in the main API binary.

See **[RabbitMQ](./rabbitmq.md)** for exchanges, queues, env vars (`RABBITMQ_URL`, persistence, Pi notes).

---

## Telegram notifier worker (consumer)

Small **Go binary** consuming from **`telegram.outbound.durable`** (or your chosen queue):

- Parses JSON payloads (minimal schema: `text`, optional `severity`, `source`, correlation id).
- Calls `POST https://api.telegram.org/bot<TOKEN>/sendMessage` with **`chat_id`** and **`text`** (Markdown/HTML only if escaped correctly).
- Implements **timeouts**, **bounded retries**, and optional **circuit breaker** if Telegram responds `429 Too Many Requests`.

**Minimum configuration**

| Secret / env | Role |
| --- | --- |
| `TELEGRAM_BOT_TOKEN` | From `@BotFather`. Stored only where the worker runs. |
| `TELEGRAM_CHAT_ID` | User or channel numeric id reachable by your bot. |
| `RABBITMQ_URL` | Consume connection string. |

---

## Producers that feed RabbitMQ

| Producer | Trigger |
| --- | --- |
| **Go API** | Explicit publish on curated **critical** branches (never per-request spam)—e.g. sustained dependency failure surfaced by application logic alongside Prometheus. |
| **Alertmanager webhook → AMQP relay** *(optional)* | Prometheus fires → Alertmanager **`webhook_configs`** **`POST`** to your relay (`http://alert-amqp-gateway:8090/` on Docker network)→ relay **publishes** one framed message per alert or bundled digest. |

Keeping **Prometheus alerting** routed through RabbitMQ aligns **one queue depth** observable under stress instead of synchronous Telegram posts that can block alert delivery.

Legacy **direct Alertmanager → Telegram webhook** bypasses buffering; migrate to **relay → RabbitMQ → worker** when demonstrating message durability.

---

## What to notify on vs noise

- Route only **`severity=critical`** / sustained SLO breaches; use `for:` in Prometheus rules before enqueueing.
- **Resolved** alerts optional in Telegram (**dedupe** by fingerprint in worker or omit resolved messages entirely for coursework).

---

## Security

- **Never commit** bot tokens or chat IDs; use host `.env` / Kubernetes **Secrets** on the edge lab host.
- **Do not** expose RabbitMQ **15672 management** publicly; bind **`127.0.0.1`** and **SSH tunnel** from the workstation, or LAN firewall only.

---

## Deployment on the edge lab host

- Run **broker + notifier + optional relay** in **Docker Compose** on the Raspberry Pi alongside the API stack.
- The worker needs **HTTPS egress** to Telegram; RabbitMQ listens **inside** the Compose network (**5672**).
- Inspect backlog: **`docker compose logs telegram-notifier`**, **`rabbitmqadmin`** or management UI queues tab.

Upstream: **[Prometheus](./prometheus.md)** rules and Alertmanager webhook configuration; **[Edge lab host](./edge-lab-host.md)** sizing (RAM for RabbitMQ Erlang VM on Pi).
