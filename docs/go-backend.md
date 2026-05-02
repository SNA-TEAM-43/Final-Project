# Go HTTP backend

The API process is the **core workload**: it serves HTTP, exposes Prometheus metrics, logs requests and errors, and uploads multipart images to S3-compatible storage. In diagrams it appears as **Go HTTP API replicas** inside Kubernetes; behind NGINX when you mimic production.

See **[Edge lab host](./edge-lab-host.md)** for the primary deployment topology (Pi / Ubuntu).

---

## Responsibilities

- **HTTP routing:** health and optional JSON APIs (`GET`), writes such as multipart upload (`POST`).
- **Durability boundary:** offload blob bytes to S3 via server-side SDK (`PutObject` or equivalent multipart API if you evolve past small images).
- **Observability:** register Prometheus collectors and middleware (request counting, histograms); emit structured logs (stdout JSON recommended).
- **Operational hooks:** semantics should match Kubernetes **readiness** and **liveness** paths (often separate).

---

## API endpoints (target contract)

These are the intended routes for the coursework stack. Adjust implementations to match exactly so NGINX, probes, and scripts stay stable.

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/health` | **Liveness:** process is up; cheap; no dependency checks (or only in-process). Return `200` with a small JSON body. |
| `GET` | `/ready` | **Readiness:** dependencies OK (e.g. S3 reachability or config); return `503` when not ready so load balancers stop sending traffic. Optional if you only use `/health` for both. |
| `GET` | `/metrics` | **Prometheus:** text exposition format; no auth in dev (lock down in real deployments). |
| `GET` | `/api/ping` | **Demo read:** deterministic JSON `{ "status": "ok" }` for dashboards and smoke tests. |
| `POST` | `/upload` | **Image upload:** `multipart/form-data` with field name `file`; store as object `{prefix}/{uuid-or-hash}{ext}` or similar; return `201` with key/URL metadata or `400/413/502` as appropriate. |

**Headers:** Preserve `Content-Type`, `Content-Length`, and multipart boundaries end-to-end; NGINX defaults usually pass them through untouched.

---

## Request and error logging

- Prefer **structured JSON** per line (`time`, `level`, `method`, `path`, `status`, `duration_ms`, optional `trace_id`).
- Log **panic recoveries** and **S3 errors** at `error`; avoid logging raw file bytes.
- Correlate with Prometheus: increments on `handler`/`method`/status buckets keep RED-style metrics coherent with log lines.

---

## Configuration (environment variables)

| Variable | Role | Example |
| --- | --- | --- |
| `HTTP_LISTEN_ADDR` | Bind address (`:8080`). | `:8080` |
| `S3_ENDPOINT_URL` | Custom endpoint for AWS SDK (MinIO, LocalStack). Empty for default AWS resolver. | `http://minio:9000` |
| `S3_REGION` | Region string. | `us-east-1` |
| `S3_BUCKET` | Target bucket. | `uploads` |
| `S3_ACCESS_KEY_ID` / `S3_SECRET_ACCESS_KEY` | Static credentials for dev **only** (in cloud use IAM / IRSA). | from `.env` |
| `S3_USE_PATH_STYLE` | `true` for many MinIO setups. | `true` |
| `UPLOAD_PREFIX` | Key prefix (`images/`). | `images/` |
| `ENV` | `dev`/`prod` toggle for verbose logs. | `dev` |

Add more as implementation grows (timeouts, max body size).

---

## Suggested repo layout for this binary

Typically `cmd/server/main.go` with handlers under `internal/api` and storage under `internal/storage`. Dockerfile builds this target.

---

## Deployment on the edge lab host

- Run inside **Docker Compose** on the host (service `api`), listening on the internal Docker port (example `8080`); expose it to the outside world **only via NGINX** unless you consciously open the API port for debugging.
- Configure `S3_*` so the container reaches **MinIO** on the same Compose network (`http://minio:9000`) or **AWS** over HTTPS from the Pi.
- After `docker compose up`, verify from the workstation: `curl http://<host-ip>:9080/health` (substitute published NGINX port).
- Operational visibility: **`docker compose logs -f api`** over **SSH**, or use the **[API endpoint table](#api-endpoints-target-contract)** from the workstation (`curl …/health`, `…/api/ping`).

See also [Edge lab host](./edge-lab-host.md), [Docker Compose](./docker-compose.md), [S3](./s3-storage.md), [Prometheus](./prometheus.md), and [Kubernetes](./kubernetes.md) for probe paths and scraping.
