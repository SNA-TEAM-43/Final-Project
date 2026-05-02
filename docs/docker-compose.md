# Docker Compose (local & demos)

Defines a **multi-container network** that runs the stack without Kubernetes: API, NGINX, Prometheus, optional MinIO, **RabbitMQ**, **Telegram notifier worker**, optional **Alertmanager‑to‑AMQP relay**, optional Alertmanager.

On the **[edge lab host](./edge-lab-host.md)**, Compose is usually the **default** orchestration (lighter than Kubernetes on a Pi).

---

## Typical services

| Service | Image / build | Role |
| --- | --- | --- |
| `api` | `build: .` target server | Go HTTP backend on `8080` internally. |
| `nginx` | `nginx:alpine` + config mount | Publishes `9080:80` (example) to host; proxies to `api`. |
| `prometheus` | `prom/prometheus` | Mount `prometheus.yml`; scrape `api:8080/metrics`. |
| `minio` | `minio/minio` | S3 API; create bucket via init container or script. |
| `rabbitmq` | `rabbitmq:3-management-alpine` | AMQP **5672** on internal network; **15672 management** bind to `127.0.0.1` when possible (**[rabbitmq.md](./rabbitmq.md)**). |
| `telegram-notifier` | `build: .` target notifier | Consumes RabbitMQ queue; calls Telegram Bot HTTP API (**[telegram.md](./telegram.md)**). |
| `alert-amqp-gateway` | `build: .` *(optional)* | HTTP receiver for **`webhook_configs`** from Alertmanager → **publish** normalized JSON to RabbitMQ same as API-critical events. |

Add **`alertmanager`** when you demo firing rules toward the **gateway** (**not** direct Telegram webhook if you showcase buffering consistently).

---

## Configuration files

| File | Purpose |
| --- | --- |
| `docker-compose.yml` | Services, networks, depends_on, published ports. |
| `.env` (gitignored) | `S3_*`, MinIO root keys, **`RABBITMQ_URL` / RabbitMQ defaults**, Telegram **worker** tokens (**not necessarily** on `api`). |
| `deploy/prometheus/prometheus.yml` | Scrape job for `api:8080`. |
| `deploy/nginx/*.conf` | Upstream and `client_max_body_size`. |

---

## Environment wiring

- Pass **`S3_*`** and **`RABBITMQ_*` / AMQP URLs** into the `api`; pass **`TELEGRAM_*` + `RABBITMQ_URL`** only on **`telegram-notifier`** (recommended split).
- Use **service DNS names** (`http://minio:9000`) inside the Compose network, not `localhost`.

---

## Common commands

```bash
docker compose up --build
docker compose ps
docker compose logs -f api
```

For production-like behaviour from your **workstation**, point the [load client](./go-load-client.md) at **`http://<edge-lab-host-ip>:9080`** (through NGINX). Use `localhost` only when the client runs on the same machine as Compose.

---

## Deploying on the edge lab host

1. Install Docker Engine + Compose plugin (Ubuntu Server on Raspberry Pi is supported **arm64**).
2. `git clone` this repository on the host (or rsync tarball); ensure `deploy/` configs and compose file paths match the repo layout you implement.
3. Create **`.env`** on the Pi (chmod `600`) with `S3_*` and MinIO keys; never sync `.env` to public branches.
4. `docker compose up -d --build` from the compose file directory; bind published ports (**9080**, optional **9090** for Prometheus) to **LAN** only or **localhost + SSH tunnel** if you expose the UI briefly.
5. Open **SSH** from the workstation when you narrate **`docker compose ps`** and **`docker compose logs`** during fault demos.

See [Edge lab host](./edge-lab-host.md) for topology and sizing notes (SD wear, Prometheus retention).
