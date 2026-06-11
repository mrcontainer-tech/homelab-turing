# ArgoCD

GitOps controller. **Currently installed manually** (raw upstream manifests +
`--insecure` flag, live version v3.0.6); only the cert, ingresses, and
ServiceMonitor in `manifests/` are GitOps-managed. This is the long-standing
TODO from the project docs.

## Planned migration to Helm (self-management)

The groundwork is in place but **deliberately not enabled**:

- `values.yaml` in this directory replicates the current config
  (`server.insecure: true`, single replicas, domain).
- A commented-out chart entry exists in
  `appsets/core-components-chart-application-set.yaml`.

### Why not enabled yet

Migrating ArgoCD to manage itself is the one change that can brick the GitOps
loop if it goes wrong. Do **not** enable it until:

1. The cluster is fully healthy (in particular: cluster DNS working, all apps
   syncing — see `analysis-report.html` from 2026-06-10).
2. You have a recovery path ready (below).

### Migration steps

1. Back up the current install state:
   `kubectl get all,cm,secret,ingress -n argocd -o yaml > argocd-backup.yaml`
2. Verify the chart version: pick the argo-helm `argo-cd` chart release that
   ships your current app version (v3.0.6 → chart 8.1.x) so the migration is a
   pure ownership change, not a version jump. Upgrade afterwards, separately.
3. Uncomment the entry in `appsets/core-components-chart-application-set.yaml`
   and set the verified version.
4. Commit and push. The release name `argocd` produces the same resource names
   as the manual install (`argocd-server`, `argocd-repo-server`, …), and
   `ServerSideApply=true` lets ArgoCD adopt the existing objects in place.
5. Watch `argocd-chart` sync. Expect a one-time rollout of all components.
6. Verify: `argocd app list`, UI reachable, and that `server.insecure` took
   effect via `configs.params` (the old container `--insecure` flag becomes
   redundant).

### Recovery

If ArgoCD breaks mid-migration, reinstall the known-good upstream manifests by
hand and remove the chart entry from the ApplicationSet:

```bash
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v3.0.6/manifests/install.yaml
```

## Note on namespace conventions

This component deploys ingress/cert resources into the `argocd` namespace,
which matches the directory name — no exception here. The ServiceMonitor
labels must carry `release: kube-prometheus-stack` to be scraped.
