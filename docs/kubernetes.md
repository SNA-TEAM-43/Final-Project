# Kubernetes

Runs the workload as **Pods** controlled by **Deployments**, exposed through **Services** (and Ingress or ingress NGINX if you mimic external HTTP). Automated **restarts** and **rolling updates** illustrate recovery scenarios.

---

## Core resources

| Resource | Responsibility |
| --- | --- |
| `Deployment` | Desired replicas, rolling update strategy, pod template (`image`, `env`, probes). |
| `Service` (ClusterIP) | Stable VIP/DNS selecting Pods by labels; NGINX or Ingress fronts this. |
| `ConfigMap` / `Secret` | Non-secret config (`S3_REGION`, bucket name); TLS; access keys (prefer external secret store in real systems). |
| `HorizontalPodAutoscaler` | Optional coursework extension for scale-under-load demos. |

---

## Probes (align with [API](./go-backend.md))

| Probe | Path | Semantics |
| --- | --- | --- |
| `livenessProbe` | `GET /health` | Restart container if failing (process wedged). |
| `readinessProbe` | `GET /ready` or second check on `/health` | Remove from Service endpoints if failing (dependencies down). |

Use **HTTP** probes with `initialDelaySeconds` tuned so short cold starts do not flap.

---

## Resource requests and limits

- Set **`requests`** near steady usage and **`limits`** above spike to throttle OOM during fault injection labs.
- “Delete pod” and “OOMKill” demos require limits to behave predictably.

---

## Typical manifest snippets

Environment from Secret:

```yaml
envFrom:
  - secretRef:
      name: api-s3
```

Prometheus scraping in-cluster: annotate Pods for scraping or expose a scrape config that targets Pod IPs — operator stacks automate this.

---

## Recovery behaviour to narrate during defence

ReplicaSet replaces failed Pods; **readiness** prevents traffic shift until uploads work again after S3 outages; NGINX sheds bad upstreams briefly while endpoints update.

---

## Deployment on the edge lab host (optional)

Kubernetes on a Pi is **optional** compared to **[Docker Compose](./docker-compose.md)**; reserve it when you explicitly need **`Deployment`/`StatefulSet`** lifecycle demos or coursework requires K8s.

- **Light distros:** **k3s** is common on Ubuntu/arm64 Pis; allocate **≥2 GB RAM** free after OS for apiserver + etcd + workloads.
- **Images:** Prefer **GHCR-published `linux/arm64`** images matching your Actions build so kubelet does not pull wrong-arch manifests.
- **Expose the API to the workstation:** **NodePort**, **LoadBalancer** (MetalLB on LAN), or **Ingress** with host IP—document the URL the [load client](./go-load-client.md) uses.
- **Kubeconfig over SSH:** `scp`/`ssh` merged config lets you **`kubectl logs`** from the workstation while SSH’d or via VPN—same observability story as Compose + logs.
- If the Pi strains under control-plane load, declare in your report that **Compose is the Pi default** and **Kubernetes manifests** validate on kind/minikube separately.

Baseline topology: **[Edge lab host](./edge-lab-host.md)**.
