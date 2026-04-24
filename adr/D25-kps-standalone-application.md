# D25: Extract kube-prometheus-stack into a Standalone ArgoCD Application

**Date**: 2026-04-24
**Status**: Proposed
**Decision**: Remove `kube-prometheus-stack` from the shared `applications` ApplicationSet and manage it as a standalone ArgoCD `Application` at `appsets/kube-prometheus-stack-chart-application.yaml`
**Context**: The shared AppSet's `ignoreDifferences` block is designed to coexist with the Kyverno image-mutation policy (D24). `kube-prometheus-stack` is excluded from that policy, so the stripping is incidental and breaks Helm charts that render new `initContainers` without pre-existing live state.

---

## Problem Statement

`appsets/applications-chart-application-set.yaml` applies this `ignoreDifferences` block to every generated Application:

```yaml
ignoreDifferences:
  - group: apps
    kind: Deployment            # also StatefulSet, DaemonSet
    jqPathExpressions:
      - .spec.template.spec.containers[].image
      - .spec.template.spec.initContainers[]?.image
```

Combined with `ServerSideApply=true` and `RespectIgnoreDifferences=true`, this strips the listed `.image` paths from the manifest ArgoCD sends to the API server. The intent: let the Kyverno `mutate-image-registry` ClusterPolicy rewrite images to `zot.mrcontainer.nl/...` at admission time without ArgoCD continuously reverting Kyverno's mutations.

This contract holds for any namespace Kyverno actually mutates. It breaks for namespaces excluded from the policy — the field is stripped but nothing fills it back in.

The `mutate-image-registry` policy excludes a set of infrastructure namespaces (argocd, cert-manager, cloudnativepg, coredns, dnscrypt-proxy, external-dns, external-secrets, **kube-prometheus-stack**, kyverno, longhorn-system, metallb-system, policy-reporter, tailscale, tekton-operator, tekton-pipelines, zot) to avoid a circular dependency: if Zot is unreachable, the observability/bootstrap stack must still be able to pull images.

### Observed failure (2026-04-24)

After the `kube-prometheus-stack` chart rendered a new `download-dashboards` initContainer (Grafana subchart 9.2.9), ArgoCD sync failed with:

```
Deployment.apps "kube-prometheus-stack-grafana" is invalid:
spec.template.spec.initContainers[0].image: Required value
```

The live Deployment had `initContainers: null`, so SSA had no prior-state `.image` to preserve. ArgoCD applied an initContainer entry with no `image` field and the API server rejected it.

The chart's main containers were unaffected because live state already held an `.image` value from earlier syncs; only the newly-introduced initContainer exposed the mismatch.

---

## Goals

- Restore reliable syncs for `kube-prometheus-stack`.
- Keep the Kyverno exclusion for `kube-prometheus-stack` (no new circular dependency on Zot for the observability stack).
- Keep the shared AppSet's `ignoreDifferences` contract intact for the apps that genuinely need it.
- Avoid custom templating in the shared AppSet for a single edge case.

---

## Options Considered

### Option 1: Extract into a standalone Application (this ADR)

Remove the `kube-prometheus-stack` element from the list generator. Add `appsets/kube-prometheus-stack-chart-application.yaml` — an ArgoCD `Application` with the same multi-source setup but **without** the `ignoreDifferences` block and **without** `RespectIgnoreDifferences`.

**Pros**:
- Preserves the invariant: "`ignoreDifferences` applies only where Kyverno is mutating."
- No templating or per-element override in the shared AppSet.
- Picked up automatically by `appsets/bootstrap.yaml`, which already watches `appsets/` non-recursively.

**Cons**:
- One more top-level YAML in `appsets/`.
- Mild duplication with the AppSet template.

### Option 2: Remove `kube-prometheus-stack` from the Kyverno exclusion list

Let Kyverno mutate the namespace's images so the stripped `.image` fields get re-filled at admission.

**Pros**:
- No AppSet changes.

**Cons**:
- Reintroduces a circular dependency: if Zot is unhealthy, Prometheus, Grafana, and Alertmanager images cannot be pulled — the tools needed to debug a Zot outage go down with Zot.
- The observability namespace was deliberately excluded for exactly this reason.

### Option 3: Template per-element `ignoreDifferences` in the AppSet

Add an optional `ignoreDifferences` field per list element and template it in.

**Pros**:
- Keeps everything in the shared AppSet.

**Cons**:
- Introduces Go-template branching in a central file for a single special case.
- Harder to read than a separate manifest.

---

## Decision

**Option 1**. The root cause is a semantic mismatch — `ignoreDifferences` in the shared AppSet exists to accommodate Kyverno mutation, and `kube-prometheus-stack` doesn't get mutated. Colocating it in the shared AppSet is incidental, not intentional. Extracting it makes the relationship explicit.

The same reasoning applies to any other Kyverno-excluded namespace that the user later migrates into the `applications/` ApplicationSet. If more cases appear, reconsider Option 3.

---

## Implementation Plan

1. Create `appsets/kube-prometheus-stack-chart-application.yaml` — a standalone `Application` keeping the existing name `kube-prometheus-stack-chart` so the destination resources continue to be tracked by the same name. No `ignoreDifferences`, no `RespectIgnoreDifferences=true`.
2. Remove the `kube-prometheus-stack` element from the list generator in `appsets/applications-chart-application-set.yaml`. Leave a short comment referencing this ADR.
3. Add D25 row to `adr/README.md`.

### Migration note

The currently-running `kube-prometheus-stack-chart` Application is owned by the `applications` ApplicationSet (`ownerReferences`) and has the `resources-finalizer.argocd.argoproj.io` finalizer. When the list element is removed, the AppSet controller will delete the generated Application; the finalizer then prunes the managed resources (Deployments, StatefulSets, Services, etc.). The new standalone Application will then re-sync everything from scratch.

Expected impact:
- Prometheus / Grafana / Alertmanager pods restart once. No PVCs exist in this namespace (storage is `emptyDir`), so the restart wipes in-memory metric history — acceptable for a homelab.
- `namespace`, cert, and ingress manifests live in `applications/kube-prometheus-stack/manifests/` and are owned by the `applications-manifests` AppSet, so they are not affected.

To avoid the brief delete-recreate cycle, optionally edit the existing Application in the cluster to remove the `resources-finalizer.argocd.argoproj.io` finalizer before merging this change. The Application will then be deleted by the AppSet controller without cascading to managed resources, and the new standalone Application will adopt them via SSA. This is a manual step intentionally left outside GitOps.

---

## Dependencies

| ADR | Relationship |
|---|---|
| D10 | Builds on: Kyverno policy enforcement — the excluded-namespace list is the reason the stripping contract breaks here. |
| D24 | Builds on: Zot registry + image-mutation policy. `kube-prometheus-stack` is excluded from that policy to avoid a circular dependency on Zot for observability. |

---

## References

- `appsets/applications-chart-application-set.yaml` — source of the shared `ignoreDifferences` contract
- `appsets/kube-prometheus-stack-chart-application.yaml` — the extracted Application
- [ArgoCD: Diffing Customization](https://argo-cd.readthedocs.io/en/stable/user-guide/diffing/)
- [ArgoCD: Server-Side Apply](https://argo-cd.readthedocs.io/en/stable/user-guide/sync-options/#server-side-apply)
