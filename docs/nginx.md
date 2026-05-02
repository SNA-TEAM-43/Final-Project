# NGINX (reverse proxy & load balancer)

Terminates inbound traffic (TLS optional), balances across **upstream** Go backends, applies **timeouts**, and forwards headers so the app sees the original scheme/host when relevant.

---

## Typical placement

```
Client → nginx:80/443 → upstream api:8080 (one or many servers)
Prometheus scrapes backends directly (/metrics); optional nginx stub_status exporter for edge metrics only.
```

---

## Core configuration knobs

Example concerns (adapt file paths to your repo, e.g. `deploy/nginx/default.conf`):

| Directive | Purpose |
| --- | --- |
| `upstream api { ... }` | Named pool of `server host:port` entries; NGINX rotates by method (`least_conn`, `round_robin`). |
| `proxy_pass http://api;` | Forward to pool. |
| `proxy_connect_timeout` / `proxy_read_timeout` | Avoid hung clients; surface pressure under slow backends. |
| `client_max_body_size` | Must allow your **largest image** plus multipart overhead. |
| `proxy_set_header Host` / `X-Forwarded-For` / `X-Forwarded-Proto` | Preserve client context for logs and redirects. |

---

## Health and metrics

- **Health checks** on NGINX upstreams are coarse; Kubernetes **readiness** on Pods is authoritative for draining.
- Optionally expose **`stub_status`** for basic connection counts; scraping **nginx-exporter** is optional coursework sugar.

Integrates with routes listed in [Go backend](./go-backend.md) (paths must match forwarding rules).

---

## Deployment on the edge lab host

- Run NGINX **on the Pi** as a Compose service with a mounted `default.conf`; publish **`:9080→80`** (example) so the **workstation** load client hits port `9080` on the Pi’s LAN IP.
- Keep TLS optional for coursework; behind home NAT prefer **HTTPS** only if you port-forward externally.
- After upstream container restarts, NGINX resumes once the Compose network DNS resolves the `api` service again—use this to narrate brief **502** spikes during rollout or kill experiments.

Topology: [Edge lab host](./edge-lab-host.md).
