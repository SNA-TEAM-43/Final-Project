# SNA Final Project — Observability & Self-Healing Demo

Demonstrate the full observability loop: **simulate failures → backend heals itself → everything is visible in Grafana**.

The system spans two servers and a local CLI client. The backend automatically resets stuck jobs (self-healing watchdog), exposes Prometheus metrics, and emits structured JSON logs. The monitoring stack collects everything into Grafana dashboards and alerts. The CLI client drives load and fault injection scenarios.

## Repository layout

```
Final-Project/
├── client/                    # Interactive Go CLI load generator
└── services/
    ├── backend/               # Go HTTP API + PostgreSQL + collectors  (git submodule)
    └── monitoring/            # Grafana + Loki + Mimir + Traefik       (git submodule)
```

## Deployment topology

```
┌─────────────────────────────────────────┐
│  Local machine                          │
│  client/  ── HTTP                       |
└─────────────────────────────────────────┘                  
                   ▼                                          
┌─────────────────────────────────────────┐      ┌──────────────────────────────┐
│  Application server                     │      │  Monitoring server           │
│                                         │      │                              │
│  services/backend/                      │      │  services/monitoring/        │
│  ├── api          :8080  ◄─── requests  │      │  ├── traefik  :80/:443       │
│  └── db           (PostgreSQL 17)       │      │  ├── grafana  :3000          │
│                                         │      │  ├── mimir    :9009  ◄───────┤
│  services/backend/scrapers/             │      │  └── loki     :3100  ◄───────┤
│  ├── alloy ── remote_write ─────────────┼─────►│                              │
│  │         ── log push ─────────-───────┼─────►│                              │
│  ├── node_exporter                      │      │                              │
│  └── postgres_exporter                  │      └──────────────────────────────┘
└─────────────────────────────────────────┘
```

## Components

| Component | Runs on | Port | What it does |
|-----------|---------|------|--------------|
| **Backend API** | Application server | 8080 | Job CRUD, load simulation, chaos endpoints, `/metrics` |
| **PostgreSQL 17** | Application server | — (internal) | Persistent storage; `pg_stat_statements` enabled |
| **Grafana Alloy** | Application server | 12345 (local) | Scrapes metrics → Mimir; ships Docker logs → Loki |
| **node_exporter** | Application server | 9100 (internal) | Host CPU, memory, disk, network metrics |
| **postgres_exporter** | Application server | 9187 (internal) | PostgreSQL runtime statistics |
| **Traefik** | Monitoring server | 80 / 443 | Reverse proxy + automatic TLS (Let's Encrypt) |
| **Grafana** | Monitoring server | 3000 / via Traefik | Dashboards, alerting, log and metrics exploration |
| **Grafana Mimir** | Monitoring server | 9009 | Long-term metrics store (PromQL) |
| **Grafana Loki** | Monitoring server | 3100 | Log aggregation (LogQL) |
| **Load client** | Local machine | — | Interactive CLI: normal load, stress, kill, chaos |

## Quick start

Initialize submodules first:

```bash
git submodule update --init --recursive
```

### 1. Monitoring server

```bash
cd services/monitoring
cp .env.example .env        # set GRAFANA_DOMAIN
docker compose up -d
```

Grafana: `https://<GRAFANA_DOMAIN>` — credentials `admin` / `admin`.

See [services/monitoring/README.md](services/monitoring/README.md) for full setup (Traefik TLS, firewall requirements).

### 2. Application server

```bash
cd services/backend
cp .env.example .env        # set DATABASE_URL and Postgres credentials
docker compose up -d --build

cd scrapers
cp .env.example .env        # set MIMIR_PUSH_URL, LOKI_PUSH_URL, POSTGRES_DSN
docker compose up -d
```

See [services/backend/README.md](services/backend/README.md) for full API reference and configuration.

### 3. Load client (local machine)

```bash
cd client
echo 'SERVER_URL=http://<application-server-ip>:8080' > .env
go run .
```

## Demo scenarios

| Client option | What happens | What to watch in Grafana |
|---------------|-------------|--------------------------|
| **1 — CPU load** | 2-second CPU burn, 2 workers | CPU usage spike |
| **2 — DB write** | 100 concurrent inserts | `db_queries_total{operation="write"}` |
| **3 — DB read** | 50 concurrent selects | `db_queries_total{operation="read"}` |
| **5 — STRESS** | CPU + DB write + DB read in parallel | `http_requests_total`, latency histogram |
| **6 — KILL** | Max-intensity load for ~5 min | All metrics at saturation |
| **7 — CHAOS** | DB writes fail for 30 s (auto-generated load) | `db_queries_total{status="failed"}` spike |
| **8 — HEAL** | Cancel chaos immediately | Failed rate drops to zero |
| **Self-healing** | Manually set a job to `processing`, wait 2 min | `healed_jobs_total` increments, log line in Loki |

## Self-healing watchdog

The backend runs a background goroutine that checks every 30 seconds for jobs stuck in `processing` for more than 2 minutes and resets them to `created`. Each reset:

- increments the `healed_jobs_total` Prometheus counter
- emits a structured log line: `{"msg":"job healed","id":N,"stuck_for_sec":M}`

Both the counter and the log are visible in Grafana (Mimir + Loki datasources).

## Metrics exposed by the backend

| Metric | Labels | Description |
|--------|--------|-------------|
| `http_requests_total` | `method`, `path`, `status` | HTTP request count |
| `http_request_duration_seconds` | `method`, `path` | Latency histogram |
| `db_queries_total` | `operation`, `status` | DB load-test operations |
| `healed_jobs_total` | — | Watchdog job resets |
