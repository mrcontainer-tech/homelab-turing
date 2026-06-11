# D27: conveyor-system runs outside GitOps as a documented exception

**Date**: 2026-06-10
**Status**: Proposed
**Decision**: We allow the OpenConveyor controller (`conveyor-system` namespace) to be deployed imperatively from its development repo via kubectl/kustomize, as a time-boxed exception to the "everything through ArgoCD" rule, with a defined adoption path.
**Context**: The cluster's stated philosophy is that Git is the single source of truth, enforced by ArgoCD ApplicationSets with auto-sync, prune, and self-heal. The OpenConveyor controller is under active development and has been running outside that model since ~2026-04-19.

---

## Problem Statement

The 2026-06-10 cluster analysis found a namespace `conveyor-system` running `conveyor-controller-manager`, applied with `kubectl apply` (kustomize-managed labels, `last-applied-configuration` annotations) and not represented anywhere in this repository. Meanwhile `applications/conveyor-tasks/` *is* GitOps-managed but contains only the task namespace and ExternalSecrets (Anthropic API key, GitHub token, Claude Code OAuth token) consumed by conveyor workloads.

This violates the repo's own rule that all cluster state flows through Git, and creates concrete risks:

- The controller is invisible to ArgoCD drift detection, the analysis tooling, and anyone reading this repo to understand the cluster.
- A cluster rebuild from this repository would silently lose the component.
- Nothing prevents the dev-repo deployment and a future GitOps adoption from fighting over the same resources.

At the same time, the controller is a fast-iterating development project: its manifests change with the code, and pinning a snapshot into this repo while iterating would either constantly lag or — worse — let ArgoCD self-heal revert in-development changes applied from the dev repo mid-iteration.

## Decision

We accept imperative deployment of `conveyor-system` from the OpenConveyor development repository **as a temporary, documented exception**, bounded as follows:

- The exception covers only the `conveyor-system` namespace. Supporting resources it depends on (the `conveyor-tasks` namespace and its ExternalSecrets) remain GitOps-managed in `applications/conveyor-tasks/`.
- This ADR is the system of record for the exception; the component must not silently expand (no additional namespaces, CRDs used by other components, or cluster-wide RBAC beyond what the controller needs).
- **Adoption trigger**: once OpenConveyor reaches a taggable/releasable state — or after 3 months (by 2026-09-10), whichever comes first — its manifests (or a pinned kustomize/Helm reference) move into `applications/` or `core-components/` and this exception ends.

## Alternatives considered

- **Option A — Vendor the live manifests into this repo now:** restores the single-source-of-truth invariant immediately. Rejected for now because ArgoCD self-heal/prune would fight the kustomize deploys from the dev repo during active iteration, reverting work-in-progress and creating confusing partial states.
- **Option B — GitOps from the dev repo directly:** point an ArgoCD Application at the OpenConveyor repository's kustomize overlay. Cleanest long-term shape and the likely adoption mechanism, but premature while the manifest layout is still churning; every refactor would break the Application.
- **Option C — do nothing (undocumented drift):** the status quo. Rejected: the component remains invisible, a rebuild loses it, and the project's GitOps guarantee quietly stops being true without anyone deciding that.

## Consequences

### Positive
- Development on OpenConveyor stays fast; no tug-of-war between ArgoCD self-heal and dev iterations.
- The exception is visible, scoped, and time-boxed instead of being silent drift.
- Dependencies (secrets, task namespace) stay GitOps-managed, so the blast radius of the exception is the controller itself.

### Negative / accepted trade-offs
- A cluster rebuild from this repo does not restore `conveyor-system`; the dev repo's deploy step is required and must be documented there.
- No drift detection or audit trail for the controller while the exception lasts.
- Time-boxed exceptions have a habit of becoming permanent; the trigger below exists to counter that.

### Follow-ups
- Add a deploy/README note in the OpenConveyor repo stating that this cluster expects `kubectl apply -k` from there, referencing this ADR. (Owner: Martijn)
- On the adoption trigger (release or 2026-09-10): create the component directory in this repo (Option B mechanism preferred — ArgoCD Application pointing at a pinned ref of the OpenConveyor repo) and mark this ADR Superseded.
- Revisit immediately if conveyor-system grows cluster-scoped RBAC or CRDs consumed by other components.

## References

- `applications/conveyor-tasks/` — GitOps-managed supporting resources
- `analysis-report.html` (2026-06-10) — finding H2, unmanaged drift
- D5 — Use ArgoCD for GitOps (the rule this ADR carves an exception from)
