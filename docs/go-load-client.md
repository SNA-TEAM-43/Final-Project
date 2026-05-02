# Go load-testing client

A separate **`cmd/loadtest`** (or equivalent) binary that drives **controlled load** against the API. In the **[edge lab host](./edge-lab-host.md)** model it runs on the **workstation**, not on the Raspberry Pi/server; traffic crosses the network to mimic **timeouts and partial connectivity**.

---

## Responsibilities

- Issue **mixed traffic**: repeated `GET` (health, ping), occasional `POST` or multipart uploads.
- Respect **timeouts** and optional **think time** between requests.
- Aggregate **latency** (avg, p50, p95, p99 where practical), **error counts**, and **timeouts** per scenario.
- Optionally write a **Markdown or JSON summary** artifact for grading or Actions upload.

---

## Configuration (CLI flags or env)

Suggested flags:

| Flag / env | Role |
| --- | --- |
| `-base-url` | Base URL pointing at the **edge lab host** (recommended: `http://<pi-ip>:9080` through NGINX). Use `localhost` only when Compose runs on the same machine as the client. |
| `-workers` | Concurrent goroutines issuing requests. |
| `-duration` | Total run length (`30s`, `5m`). |
| `-rps` cap | Optional throttle for reproducible demos. |
| `-upload-file` path | Same small PNG/JPEG used repeatedly for multipart tests. |

---

## Scenarios relevant to demos

| Scenario | What to show |
| --- | --- |
| Steady read | Burst `GET /api/ping` and `/health`; watch Prometheus histograms flatten after warm-up. |
| Upload mix | Periodic `POST /upload`; demonstrates S3 path under load and error spikes if storage degrades. |
| Through NGINX vs direct | Same client against `:9080` and `:8080` to illustrate hop overhead. |

Outputs pair well with [Prometheus](./prometheus.md) graphs on the host and—with Kubernetes—with events when you restart pods mid-run.

---

## Deployment (workstation, not edge lab host)

- Build or download the **`loadtest`** binary on your laptop; no container on the Pi is required unless you deliberately choose remote execution.
- Point `-base-url` at the NGINX-published port on the **edge lab host** so exercised paths match reality (compression, buffering, timeouts at the proxy).
- For grading or CI artifacts, archive the textual summary locally or upload from an Actions job—but the **interesting** WAN/LAN runs happen from your workstation against the Pi.
