# Documentation index

Topic-specific docs for each part of the stack. Paths are stable for linking from coursework or reports.

**Deployment model:** **[Edge lab host](./edge-lab-host.md)** (for example Raspberry Pi with Ubuntu)—runs all server-side stack; the **[Go load-testing client](./go-load-client.md)** runs on your **workstation** and calls the host over the network.

| Doc | Covers |
| --- | ------ |
| [Edge lab host](./edge-lab-host.md) | Pi-class server topology, SSH/API operations, deploying each stack part on the host |
| [Go HTTP backend](./go-backend.md) | REST API surface, Prometheus metrics path, structured logging, backend env vars |
| [Go load-testing client](./go-load-client.md) | Synthetic traffic, tuning concurrency, interpreting output |
| [S3-compatible storage](./s3-storage.md) | Bucket layout, IAM-style considerations, MinIO vs AWS |
| [NGINX](./nginx.md) | Reverse proxy, load balancing, timeouts, forwarding `Host`/`X-Forwarded-*` |
| [Docker Compose](./docker-compose.md) | Local stack services, volumes, `.env`, typical commands |
| [Prometheus](./prometheus.md) | Scrape configs, scrape targets, alerting concepts, Grafana |
| [Kubernetes](./kubernetes.md) | Deployments, Services, probes, resources, rollout recovery |
| [GitHub Actions](./github-actions.md) | CI/CD workflows, registry, secrets, deploy + reports |
| [Telegram notifications](./telegram.md) | Alertmanager receivers, webhook-style bots, what to notify on |

Upstream overview and diagrams remain in the [main README](../README.md).
