# ArgoCD

GitOps controller, **self-managed via the argo-helm chart** (entry in
`appsets/core-components-chart-application-set.yaml`, values in this
directory). Adopted in place from the original manual `install.yaml` install
(v3.0.6, `--insecure` flag) on 2026-06-11 — chart 8.1.2 ships exactly v3.0.6,
so the adoption was a pure ownership change. The cert, ingresses, and
ServiceMonitor live in `manifests/` (separate `argocd-manifest` app).

## Upgrading

Bump `version:` in the appset entry — **one minor at a time** (ArgoCD does
not support skipping minors). Before each step, read the upstream upgrade
notes (`docs/operator-manual/upgrading/`) and wait for the previous rollout
to be fully healthy.

Ladder from v3.0.6 to v3.4.3 (chart→app, verified against the argo-helm
index on 2026-06-11):

| Chart  | Ships  | Notes                                            |
|--------|--------|--------------------------------------------------|
| 8.1.2  | v3.0.6 | Adoption baseline                                |
| 9.0.6  | v3.1.9 | Chart 8→9 boundary is benign (cmd-params only)   |
| 9.3.7  | v3.2.6 |                                                  |
| 9.5.11 | v3.3.9 |                                                  |
| 9.5.20 | v3.4.3 | Latest as of 2026-06-11                          |

Expect the `argocd-chart` app to briefly show Unknown/Progressing while
ArgoCD's own server/controller restart mid-sync — it converges on its own.

## Adoption notes (what was checked, 2026-06-11)

- Release name `argocd` renders the same resource names as the manual
  install (`argocd-server`, `argocd-repo-server`, StatefulSet
  `argocd-application-controller`, …); `ServerSideApply=true` adopts them
  in place.
- `configs.secret.createSecret: false` — the chart must never manage
  `argocd-secret` (live one holds `admin.password`, `server.secretkey`,
  TLS material; the chart would render it empty).
- `server.insecure: true` moved to `argocd-cmd-params-cm` via
  `configs.params`, replacing the container `--insecure` flag. TLS still
  terminates at the ingress (cert-manager).
- One-time image rollouts on adoption: dex v2.41.1→v2.43.1, redis
  7.2.7→7.2.8 (ecr-public mirror).
- The `argocd-redis` auth secret already existed; the chart's
  `argocd-redis-secret-init` Job only creates it if absent.
- Pre-adoption backup: `~/argocd-backup-2026-06-11.yaml` (full namespace
  dump, kept outside the repo).

## Recovery

If ArgoCD breaks badly enough that it cannot sync itself, reinstall the
last-good upstream manifests by hand and re-comment the chart entry in the
ApplicationSet:

```bash
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v3.0.6/manifests/install.yaml
```

(Replace the version with the last one that was healthy. GitOps-level
rollback — re-pinning the previous chart version in the appset — is
preferred whenever ArgoCD can still sync.)

## Note on namespace conventions

This component deploys ingress/cert resources into the `argocd` namespace,
which matches the directory name — no exception here. The ServiceMonitor
labels must carry `release: kube-prometheus-stack` to be scraped.
