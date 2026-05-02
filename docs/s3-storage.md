# S3-compatible object storage

Stores **binary image objects** for `POST /upload`. The backend should not persist uploads on Pod-local disk beyond streaming buffers.

---

## Object key layout

- Recommended pattern: `{UPLOAD_PREFIX}{timestamp-or-uuid}{original-extension}` — e.g. `images/ab12cd34.jpg`.
- Avoid user-controlled full keys to reduce path injection and listing surprises.

---

## AWS S3 vs MinIO (local)

| Concern | AWS S3 | MinIO (Compose) |
| --- | --- | --- |
| Endpoint | Default AWS | Set `S3_ENDPOINT_URL` (e.g. `http://minio:9000`) |
| Path-style | Virtual-hosted default | Often `S3_USE_PATH_STYLE=true` |
| Credentials | IAM roles, keys via secrets | `MINIO_ROOT_USER` / `MINIO_ROOT_PASSWORD` → map to app env |
| TLS | Yes in prod | HTTP acceptable on internal Docker network only |

---

## Bucket preparation

1. Create bucket once (Compose init script, `mc`, or IaC).
2. For coursework, bucket policy often **allows app identity only**, not anonymous public uploads.
3. CORS applies if browsers upload directly; with **server-side upload** via your API only, bucket CORS is less critical than **application** upload limits (`MaxMultipartMemory`, max size).

---

## Failure modes observable in Prometheus

- **Denied / wrong creds:** 5xx or 403 surfaced by your handler; increments on server error counters.
- **Timeouts:** Rising latency histogram without success count increase.

Wire credentials via [Docker Compose](./docker-compose.md) `.env` or [Kubernetes Secrets](./kubernetes.md); never commit real secrets.

---

## Deployment on the edge lab host

- **MinIO in Compose:** store data on a **volume** with enough **USB/SSD** space if the Pi microSD is small; point the API at `http://minio:9000` with path-style access as needed.
- **AWS S3:** the Pi only needs **outbound HTTPS**; credentials live in host `.env` or secrets. Latency to the region becomes part of the demo (and failure if the internet path drops).
- Bucket creation: one-time via `mc` on the host, init container, or manual console.

See [Edge lab host](./edge-lab-host.md).
