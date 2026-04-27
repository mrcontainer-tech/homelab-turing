# D26: Sync-Wave Tiering for ArgoCD AppSets

**Date**: 2026-04-26
**Status**: Proposed
**Decision**: Annotate every AppSet in `appsets/` with `argocd.argoproj.io/sync-wave` to encode a 3-tier deploy order — foundation (0), platform (10), workloads (20). Propagate the annotation through each AppSet's `template.metadata` so generated Applications carry the wave too.
**Context**: After D25 split out infrastructure-tier charts, the repo has three conceptual deploy tiers but no explicit ordering between them. ArgoCD currently relies on retry-with-self-heal to converge, producing noisy first-time bootstraps and noisy sync history.

---

## Problem Statement

Without explicit ordering, ArgoCD does the obvious thing: it tries to sync everything roughly in parallel, fails the things whose prerequisites don't exist yet (workload Application X tries to apply a CRD that's about to be installed by foundation chart Y), and self-heals on the next reconcile loop. Eventually everything converges. The mechanism is correct but loud — the sync history fills with transient errors, and a cold bootstrap takes longer than necessary because dependent resources retry on their own clock instead of waiting for prerequisites.

The repo already has a natural three-tier shape:

| Tier | Contents | Why this tier |
|---|---|---|
| Foundation (wave 0) | argocd, cert-manager, metallb, longhorn, coredns, dnscrypt-proxy, external-dns, external-secrets, tailscale, cloudnativepg, tekton-operator | Cluster prerequisites: certs, storage, networking, secret sync, registry-side operators. Nothing else can deploy without these. |
| Platform (wave 10) | kube-prometheus-stack, zot, kyverno, policy-reporter | Cluster-wide services that workloads depend on (admission, observability, registry). All four are in Kyverno-excluded namespaces (the structural reason behind D25). |
| Workloads (wave 20) | media-server, mealie, miniflux, linkding, homepage, conveyor-tasks, tekton-pipelines | User-facing applications. Depend on the platform being up. |

---

## Decision

Add `argocd.argoproj.io/sync-wave` annotations to each AppSet manifest in `appsets/`, in two places per file:

1. `metadata.annotations` on the AppSet itself — orders when the bootstrap Application creates/updates the AppSet during its sync.
2. `spec.template.metadata.annotations` — propagates the wave to every generated Application as a metadata field. Visible to tooling, useful for any future progressive-sync work.

Wave assignment with intentional gaps to allow future insertions:

| Wave | AppSets |
|---|---|
| 0 | `core-components`, `core-components-manifests` |
| 10 | `applications-infrastructure`, `applications` |
| 20 | `applications-manifests` |

### Why `applications` is at wave 10, not 20

The current contents of `applications-chart-application-set.yaml` are `kyverno` and `policy-reporter` — both platform-tier policy infrastructure, neither workloads. The naming asymmetry (`applications-chart` carrying platform charts) is honest about today's state and was preserved deliberately to keep this round zero-disruption. A future consolidation pass would move those two entries into `applications-infrastructure` and delete the now-empty `applications-chart-application-set.yaml`. Until that happens, the wave annotation reflects the actual deploy order, not the AppSet's name.

---

## What sync-waves do here, and what they don't

ArgoCD honors `sync-wave` at sync time. Within a single Application's sync, ArgoCD waits for every resource in wave N to reach `Synced + Healthy` before starting wave N+1.

**Bootstrap → AppSets**: `appsets/bootstrap.yaml` defines the parent Application that syncs every manifest in `appsets/` non-recursively. Sync-waves on those AppSet manifests order the bootstrap's own sync — foundation AppSets get applied first, then platform, then workloads.

**AppSet template → generated Applications**: waves on `template.metadata.annotations` propagate to every generated Application. These don't strictly gate the bootstrap (Applications are owned by the AppSet via `ownerReferences`, not by the bootstrap), but they're visible on every Application (`kubectl get application -o jsonpath=…`) and let tooling reason about tiers.

### Known limitation

AppSets do not carry a custom health check. From the bootstrap's perspective, an AppSet typically reports Healthy as soon as the API server accepts it — *before* its generated Applications materialize and reconcile. So sync-wave gating at the bootstrap level guarantees only that platform AppSets are *applied* after foundation AppSets, not that the foundation Applications they generate have all reached Healthy.

In practice this is fine. ArgoCD's Application controller still retries with self-heal, so even a platform Application that races ahead of its foundation prerequisite will fail once and recover. The benefit of waves here is to cut the volume of those retry storms, not to enforce a strict ordering contract.

If strict gating ever becomes necessary (e.g. before a Day-2 cluster rebuild), the upgrade path is one of:
- Progressive sync via `applicationSet.spec.strategy.type: RollingSync`.
- App-of-apps wrappers per tier with explicit health checks.
- Sync hooks on a sentinel resource per tier.

None are needed today.

---

## Implementation

Five existing AppSet manifests gain two annotation blocks each (one on the AppSet, one on the template). One new ADR (this file). One row added to `adr/README.md`.

| File | AppSet wave | Template wave |
|---|---|---|
| `appsets/core-components-chart-application-set.yaml` | 0 | 0 |
| `appsets/core-components-manifest-application-set.yaml` | 0 | 0 |
| `appsets/applications-infrastructure-chart-application-set.yaml` | 10 | 10 |
| `appsets/applications-chart-application-set.yaml` | 10 | 10 |
| `appsets/applications-manifest-application-set.yaml` | 20 | 20 |

### Outage assessment

**None expected.** Each change is an additive metadata annotation. The bootstrap detects the AppSet manifest changed, applies the new manifest via SSA. The AppSet's `spec` is structurally unchanged, so generated Applications are not regenerated. The annotation propagation to existing Applications is a metadata-only diff and applies via SSA without recreating any of their managed resources.

The only behavioral change is the order of future syncs — first sync after this lands runs strictly foundation → platform → workloads instead of all-in-parallel.

---

## Dependencies

| ADR | Relationship |
|---|---|
| D3 | Builds on: original 2-tier AppSet split (core-components vs applications). This ADR formalizes the implicit 3rd tier that emerged after D24 + D25. |
| D25 | Builds on: introduced the platform tier as a distinct AppSet. This ADR encodes the resulting order. |

---

## References

- [ArgoCD: Sync Waves](https://argo-cd.readthedocs.io/en/stable/user-guide/sync-waves/)
- [ArgoCD: ApplicationSet Progressive Syncs](https://argo-cd.readthedocs.io/en/stable/operator-manual/applicationset/Progressive-Syncs/) — the upgrade path if strict gating is ever needed
- D25: Split infrastructure-tier charts into a dedicated ApplicationSet
