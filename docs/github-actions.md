# GitHub Actions (CI/CD & reporting)

Automation running on **`push`, `pull_request`, `workflow_dispatch`** (tags optional for releases). Validates code, publishes **OCI images**, optional deploy manifests, attaches **reports** such as tests or load-test summaries.

---

## Typical workflows

| Workflow file | Responsibility |
| --- | --- |
| `ci.yml` | `go test ./...`, `golangci-lint` optional, vet, build binaries. |
| `docker-publish.yml` | `docker/build-push-action` → GHCR (`ghcr.io/<org>/<repo>:sha`). |
| `deploy.yml` (optional gated) | `kubectl apply`/Helm against a cluster secret `KUBECONFIG` or cloud auth. |

Use **workflow concurrency** (`concurrency:`) to cancel outdated runs when branches update quickly.

---

## Secrets (repository settings)

| Secret | Used for |
| --- | --- |
| `REGISTRY_TOKEN` / `GITHUB_TOKEN` | Push to GHCR (built-in token often sufficient for same repo). |
| `KUBE_CONFIG` or cloud OIDC | Cluster deploy job. |
| `TELEGRAM_BOT_TOKEN` | Only if pipeline posts deploy status (not required for Alertmanager path). |

Never echo secrets in logs; use `::add-mask::` for derived values in debug steps if needed.

---

## Artifacts and reports

- Upload **JUnit** or plain test output with `actions/upload-artifact`.
- Run [load client](./go-load-client.md) in a scheduled job against a **staging URL**; save summary **Markdown** as artifact for grading evidence.

---

## Branch protection alignment

Require **CI pass** before merge; optional required reviewers for manifests touching production namespaces.

---

## Relation to the edge lab host

- Actions runs on **GitHub-hosted runners** (or a self-hosted runner if you add one). The **Raspberry Pi / edge lab host** is the **deployment target**, not the default CI machine.
- **Build `linux/arm64` images** (and optionally `amd64`) so `docker compose pull` on the Pi uses a native image without QEMU surprises.
- **Deliver to the host** by: registry pull + `docker compose up`, **rsync** of compose files + `docker compose build` on device, or **`kubectl apply`** over SSH with a kubeconfig—pick one story and document it in your report.
- The [load-testing client](./go-load-client.md) is **not** started by Actions for the main “WAN path” narrative unless you add a scheduled job against a **public** URL; the default lab is **workstation → Pi**.

See [Edge lab host](./edge-lab-host.md).
