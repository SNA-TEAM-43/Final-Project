# Prometheus (metrics & alerting)

Pull-based monitoring: scrapes the Go **`/metrics`** endpoint on an interval, stores time series, evaluates **recording** and **alert** rules.

---

## Scrape configuration (`prometheus.yml`)

Minimal job targeting the API Pod or Compose service:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: api
    static_configs:
      - targets:
          - api:8080           # Compose service name
    metrics_path: /metrics

  # Optional: nginx stub_status exporter or node exporter for demos.
```

In Kubernetes you often rely on [**ServiceMonitor**](https://github.com/prometheus-operator/prometheus-operator) (operator) **or** static config + DNS to the ClusterIP.

---

## Metric families to expose from Go

| Family | Typical names | Purpose |
| --- | --- | --- |
| Counter | `http_requests_total{method,handler,status}` | RED “rate / errors”. |
| Histogram | `http_request_duration_seconds_bucket` | Latency SLIs. |
| Gauge | Optional in-flight / dependency health | Saturation probes. |

Process metrics (`promhttp`) add Go runtime gauges automatically when using the official registry.

---

## Alerting pipeline

```
Prometheus (rules) → Alertmanager → receivers (PagerDuty, webhook, Telegram)
```

- Configure **recording rules** only if you normalize complex queries often.
- **Alert rules**: e.g. `rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.05`.
- Silence maintenance windows via Alertmanager UI if you expose it locally.

Pairs with structured logs documented in [Go backend](./go-backend.md) (metrics for rates, logs for forensic detail).

---

## Deployment on the edge lab host

- Prometheus runs as a **Compose** service scraping **`api:8080/metrics`** on the Docker network (no WAN hop between Prometheus and API).
- Bind the Prometheus UI to **`127.0.0.1:9090`** on the Pi and use **`ssh -L 9090:127.0.0.1:9090 user@<host>`** from the workstation to view graphs safely.
- Shorten **TSDB retention** on SD-card Pi installs (`--storage.tsdb.retention.time`) to limit disk IO and wear.
- Place **alert rule files** in a mounted volume; **Alertmanager** can be a sibling container for [Telegram](./telegram.md) demos.

See [Edge lab host](./edge-lab-host.md) and [Docker Compose](./docker-compose.md).
