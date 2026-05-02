# Telegram (critical notifications)

Delivers **high-signal** human alerts when Prometheus rules fire via **Alertmanager**, or optionally when Actions reports **deploy failure**.

---

## Alertmanager receiver (recommended path)

Configure a **Webhook** receiver that targets either:

1. Telegram **Bot HTTP API**: `POST https://api.telegram.org/bot<token>/sendMessage` with JSON `chat_id` and formatted `text` (Alertmanager templating constructs the body), **or**
2. A thin **relay service** inside the cluster that validates Alertmanager HMAC/signing and forwards to Telegram (cleaner templating).

**Minimum pieces**

- Telegram **bot token** (`TELEGRAM_BOT_TOKEN`) from `@BotFather`.
- **chat id** (user or channel) reachable by the bot.
- Alertmanager `route matchers` restricting noise (only `severity=critical`, only `alertname`.

---

## What to notify on vs what to badge

Notify on persisted **burn rates** or prolonged **upstream down** (`for: 10m`). Avoid Telegram on every flaky single probe failure.

---

## Security

- Tokens live in Kubernetes **Secrets** or GitHub Secrets for pipeline-only summaries.
- Do not expose `sendMessage` to the open internet without auth on the shim.

Integrates upstream with rule examples in [Prometheus](./prometheus.md).

---

## Deployment on the edge lab host

- **Alertmanager** as a Compose (or k3s) container on the **Pi** must reach **`https://api.telegram.org`** outbound; no inbound port to Telegram is required.
- Store **bot token** and chat configuration in a **host-only** file or Docker secret; reload Alertmanager after edits.
- If the Pi loses internet, alerts queue or fail—use that as a **degraded monitoring** talking point while the API may still serve LAN clients.

See [Edge lab host](./edge-lab-host.md) and [Prometheus](./prometheus.md).
