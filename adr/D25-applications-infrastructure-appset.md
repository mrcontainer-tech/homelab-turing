# D25: Split Infrastructure-Tier Charts into a Dedicated ApplicationSet

**Date**: 2026-04-26
**Status**: Proposed
**Decision**: Manage infrastructure-tier charts (`kube-prometheus-stack`, `zot`) via a new `applications-infrastructure` ApplicationSet at `appsets/applications-infrastructure-chart-application-set.yaml`, separate from the general-purpose `applications` ApplicationSet.
**Context**: The shared `applications` AppSet's `ignoreDifferences` block is designed to coexist with the Kyverno `mutate-image-registry` ClusterPolicy. Some charts live in namespaces that are excluded from that Kyverno policy, so the contract has no fill-in and the AppSet's stripping is actively harmful for those charts.

This ADR supersedes an earlier draft of D25 ("Extract kube-prometheus-stack into a Standalone ArgoCD Application", committed in `1485cd0`) and an unmerged D26 ("Extract Zot into a Standalone ArgoCD Application"). Both proposed one-off standalone Application manifests; this revision generalizes them into an AppSet split.

---

## Problem Statement

`appsets/applications-chart-application-set.yaml` applies this `ignoreDifferences` block to every Application it generates:

```yaml
ignoreDifferences:
  - group: apps
    kind: Deployment            # also StatefulSet, DaemonSet
    jqPathExpressions:
      - .spec.template.spec.containers[].image
      - .spec.template.spec.initContainers[]?.image
```

Combined with `ServerSideApply=true` and `RespectIgnoreDifferences=true`, this strips the listed `.image` paths from the manifest ArgoCD sends to the API server. The Kyverno `mutate-image-registry` ClusterPolicy then fills them back in at admission time, rewriting them to `zot.mrcontainer.nl/...` paths. The end result: every workload pulls through Zot for caching and Trivy scanning, and ArgoCD doesn't continuously revert Kyverno's mutations.

The Kyverno policy excludes a set of infrastructure namespaces — `argocd`, `cert-manager`, `cloudnativepg`, `coredns`, `dnscrypt-proxy`, `external-dns`, `external-secrets`, `kube-prometheus-stack`, `kyverno`, `longhorn-system`, `metallb-system`, `policy-reporter`, `tailscale`, `tekton-operator`, `tekton-pipelines`, `zot` — to avoid circular dependencies (e.g. Zot cannot pull its own image through itself; observability cannot depend on the registry it monitors).

For charts in those excluded namespaces, the strip-and-fill contract breaks. Concrete failures:

- **`kube-prometheus-stack` (2026-04-24)**: chart 75.7.0 began rendering a `download-dashboards` initContainer. The live Deployment had no prior `initContainers[0].image` value for SSA to preserve, so ArgoCD applied a Deployment with an empty `.image` field. The API server rejected: `spec.template.spec.initContainers[0].image: Required value`.
- **`zot` (2026-04-26)**: the StatefulSet has been showing `OutOfSync` continuously. The live container image (`ghcr.io/project-zot/zot:v2.1.15`) is preserved across applies because SSA keeps prior values for ignored fields, but the desired-vs-live diff never converges, leaving the chart Application permanently `OutOfSync`.

---

## Goals

- Restore reliable syncs for charts whose namespaces are Kyverno-excluded.
- Keep the Kyverno exclusions intact (no new circular dependencies on Zot).
- Keep the strip-and-fill contract intact for every other chart that genuinely needs it.
- Avoid templating an exception per chart in the shared AppSet.
- Keep "infrastructure-tier" and "ordinary application" management visibly distinct in the repo.

---

## Options Considered

### Option 1: Per-chart standalone Applications (the original D25 / D26)

Drop each affected chart into a hand-written `Application` manifest in `appsets/`. Bootstrap (which already watches `appsets/` non-recursively) picks them up.

**Pros**:
- Smallest diff for the first one.
- Each Application's spec is fully visible in its own file.

**Cons**:
- Each new Kyverno-excluded chart adds a new file and (under the original convention) a new ADR.
- The repo grows two parallel patterns: AppSet-generated Applications and hand-written ones. Readers have to learn both.
- Hand-written Applications don't benefit from the AppSet template — config drift between similar Applications is easy to introduce.

### Option 2: A second ApplicationSet (this ADR)

Add `appsets/applications-infrastructure-chart-application-set.yaml`. Same template as the `applications` AppSet, minus the `ignoreDifferences` block and `RespectIgnoreDifferences=true` sync option. Charts whose namespaces are Kyverno-excluded register as elements there.

**Pros**:
- One pattern: every chart is an element in some AppSet. A new infrastructure-tier chart is a 5-line list addition.
- The AppSet's existence and name (`applications-infrastructure`) document the contract — readers don't need to dig into Kyverno excludes to know why a chart is split off.
- Mirrors the existing repo convention (`core-components-chart-...`, `applications-chart-...`).

**Cons**:
- ~30 lines of templating boilerplate are duplicated between the two AppSets. (The duplication is mechanical; the only meaningful difference is the absence of `ignoreDifferences` and `RespectIgnoreDifferences`.)
- For 2 entries today the boilerplate cost is more than the savings; the pattern starts to clearly pay off at 3+ entries.

### Option 3: Per-element `ignoreDifferences` inside the shared AppSet

Add an optional `ignoreDifferences` field per list element in `applications-chart-application-set.yaml` and template it in.

**Pros**:
- Single AppSet, single source of truth.

**Cons**:
- Go-template branching for what is conceptually a binary trait ("is this chart Kyverno-mutated?"). Harder to scan.
- Subtle to get right (rendering an empty `ignoreDifferences` is not the same as omitting it).

---

## Decision

**Option 2**. The AppSet split mirrors the existing `applications-` / `core-components-` axis the repo already uses, and the AppSet's name is its own documentation. Per-chart standalone Applications (Option 1) work in isolation but don't generalize: each new infrastructure-tier chart costs a new file and a new explanatory comment. The boilerplate duplication is a one-time cost; adding a new element is one list entry.

Option 3 (per-element override in the shared AppSet) is rejected because the trait being expressed — "this chart's namespace is Kyverno-excluded, therefore the whole strip-and-fill contract is wrong for it" — is binary and structural, not a per-chart configuration knob.

---

## Implementation Plan

1. Create `appsets/applications-infrastructure-chart-application-set.yaml` with the `applications-infrastructure` ApplicationSet. Same template as `applications`, minus `ignoreDifferences` and `RespectIgnoreDifferences=true`. Initial elements: `kube-prometheus-stack` and `zot`.
2. Remove `kube-prometheus-stack` and `zot` entries from the list generator in `appsets/applications-chart-application-set.yaml`. Replace with a comment pointing to the infrastructure AppSet and this ADR.
3. Delete `appsets/kube-prometheus-stack-chart-application.yaml` (the standalone Application from the now-superseded earlier D25) and the unmerged `appsets/zot-chart-application.yaml` (the equivalent from D26).
4. Delete `adr/D26-zot-standalone-application.md` and rename `adr/D25-kps-standalone-application.md` to `adr/D25-applications-infrastructure-appset.md`, replacing its contents with this file.

### Migration note

The cluster currently has:
- `kube-prometheus-stack-chart` Application — standalone, from the previous D25 (committed `1485cd0`). No `ownerReferences`.
- `zot-chart` Application — owned by the `applications` ApplicationSet, with `resources-finalizer.argocd.argoproj.io`.

When this PR merges, both Applications get replaced by ApplicationSet-generated Applications of the same name owned by `applications-infrastructure`. Each replacement goes through the standard ArgoCD finalizer-driven cascade: the existing Application is deleted, its managed resources are pruned, and the new Application re-syncs them.

Concrete impact:
- **kube-prometheus-stack**: Prometheus / Grafana / Alertmanager pods restart once. No PVCs in this namespace (storage is `emptyDir`), so in-memory metric history is wiped. Acceptable for a homelab.
- **zot**: StatefulSet is deleted and recreated. The PVC `zot-chart-pvc-zot-chart-0` (10 Gi Longhorn) is **not** ArgoCD-tracked — StatefulSets do not delete their PVCs on StatefulSet deletion — so the new StatefulSet reattaches by name and the cached image data in `/var/lib/registry` survives. Registry is briefly unavailable (typically <90 s while Longhorn re-attaches the volume). Most cluster workloads have already cached their images on their nodes, so restarting pods do not need to round-trip through Zot. Cluster infra (argocd, cert-manager, etc.) is unaffected — those namespaces are also Kyverno-excluded.

To avoid the restart cycle, optionally remove the `resources-finalizer.argocd.argoproj.io` finalizer from each live Application before merging — known caveat: the AppSet controller may re-add the finalizer on next reconcile, falling back to the cascade behavior.

---

## Dependencies

| ADR | Relationship |
|---|---|
| D10 | Builds on: Kyverno policy enforcement. The excluded-namespace list in `mutate-image-registry` is the structural reason this AppSet split exists. |
| D24 | Builds on: Zot registry + image-mutation policy. The Kyverno exclusions for observability and the registry itself motivate this AppSet. |

Supersedes the earlier draft of D25 ("Extract kube-prometheus-stack into a Standalone ArgoCD Application", commit `1485cd0`) and the unmerged D26 ("Extract Zot into a Standalone ArgoCD Application").

---

## References

- `appsets/applications-chart-application-set.yaml` — the general AppSet that strips images for Kyverno mutation
- `appsets/applications-infrastructure-chart-application-set.yaml` — this ADR's deliverable
- [ArgoCD: Diffing Customization](https://argo-cd.readthedocs.io/en/stable/user-guide/diffing/)
- [ArgoCD: Server-Side Apply](https://argo-cd.readthedocs.io/en/stable/user-guide/sync-options/#server-side-apply)
