# D19: CI System for Homelab

**Date**: 2026-04-03
**Status**: Accepted
**Decision**: Tekton — Kubernetes-native CI/CD (CNCF Incubating)
**Context**: No CI system currently exists in the cluster; images are built externally

---

## Problem Statement

The cluster has no CI pipeline for building container images, running tests, or pushing artifacts to Harbor. Currently:
- Knative function images are built locally or externally
- No automated build-on-push workflow exists
- Harbor acts as a pull-through cache and registry, but nothing pushes to it automatically
- The planned Talos migration (D17) adds urgency — machine configs and images may need automated validation

A CI system running on the homelab would complete the GitOps loop: code push → build → test → push to Harbor → ArgoCD deploys.

## Goals

- Build container images on ARM64 (Raspberry Pi CM4) nodes
- Push built images to Harbor registry
- Kubernetes-native or Kubernetes-compatible
- Low resource footprint (Raspberry Pi constraints)
- GitOps-friendly (pipeline-as-code)
- Self-hosted (no external service dependency)

---

## Options Considered

### Option 1: Tekton

Kubernetes-native CI/CD framework using CRDs. CNCF Incubating project.

**Architecture**:
```
Git push → Tekton EventListener → PipelineRun → TaskRuns (build, test, push) → Harbor
```

**Pros**:
- Kubernetes-native: pipelines are CRDs, runs are Pods
- CNCF Incubating — aligns with project philosophy
- Full ARM64 support with documented multi-arch builds
- Excellent GitOps fit — pipeline definitions stored in Git
- Reusable task catalog (Tekton Hub)
- Scales naturally with Kubernetes
- Supports Kaniko for rootless image builds

**Cons**:
- Higher resource footprint than lightweight alternatives
- Steep learning curve — assembling pipelines from primitives
- More YAML to manage than simpler CI tools
- No built-in UI (needs Tekton Dashboard, additional resource cost)

---

### Option 2: Woodpecker CI

Lightweight, container-native CI engine. Community fork of Drone CI.

**Architecture**:
```
Git push → Woodpecker Server → Agent → Container pipeline steps → Harbor
```

**Pros**:
- Very lightweight: ~100MB server (idle), ~30MB agent (idle)
- Full ARM64 native support
- Simple YAML pipeline syntax (`.woodpecker.yaml`)
- SQLite backend by default (zero additional infrastructure)
- Easy setup with excellent homelab documentation
- Docker-in-Docker for image builds
- Can run agents on Kubernetes

**Cons**:
- Not Kubernetes-native (server-agent architecture)
- Not a CNCF project
- Requires a Git forge integration (Gitea, Forgejo, GitHub, GitLab)
- Smaller community than Tekton

---

### Option 3: Drone CI

Container-native CI platform, predecessor to Woodpecker. Maintained by Harness.

**Pros**:
- Proven at scale, actively maintained by Harness
- Full ARM64 support (official ARM images)
- Kubernetes runner backend available
- Simple YAML pipeline syntax

**Cons**:
- Controlled by Harness (commercial company)
- Community edition has limitations
- Woodpecker is the community-driven fork with similar features
- Less homelab-focused documentation

---

### Option 4: GitHub Actions Self-Hosted Runners

Run GitHub Actions workflows on local ARM64 hardware.

**Pros**:
- Familiar GitHub Actions workflow syntax
- Full ARM64 runner support
- Free for self-hosted runners
- Large ecosystem of actions

**Cons**:
- Hard dependency on GitHub.com connectivity
- Not self-hosted for the control plane (only the runner)
- Not Kubernetes-native
- Tightly coupled to GitHub — not portable
- Runner runs as OS service, not as Kubernetes workload

---

### Option 5: Forgejo Actions

Self-hosted CI integrated with Forgejo (GitHub Actions-compatible syntax).

**Pros**:
- Fully self-hosted (git + CI in one platform)
- GitHub Actions-compatible workflow syntax
- ARM64 runner support
- No external dependencies
- Docker/LXC/shell backends for runner isolation

**Cons**:
- Requires running Forgejo as Git forge (additional infrastructure)
- Moderate resource usage
- Smaller ecosystem than GitHub Actions
- Would need to mirror/replace current GitHub repository workflow

---

## Comparison Matrix

| Criteria               | Tekton       | Woodpecker   | Drone        | GH Actions   | Forgejo      |
|------------------------|--------------|--------------|--------------|--------------|--------------|
| ARM64 Native           | Yes          | Yes          | Yes          | Yes          | Yes          |
| Kubernetes Native      | Yes (CRDs)   | No (agents)  | Partial      | No           | No           |
| Resource Footprint     | Medium-High  | Very Low     | Low          | Varies       | Moderate     |
| Setup Complexity       | High         | Low          | Moderate     | Moderate     | Moderate     |
| CNCF Status            | Incubating   | None         | None         | None         | None         |
| GitOps Friendly        | Excellent    | Good         | Good         | Moderate     | Good         |
| Self-Hosted            | Full         | Full         | Full         | Runner only  | Full         |
| Image Build Support    | Kaniko       | DinD/Podman  | DinD         | Docker       | DinD         |
| Harbor Integration     | Yes          | Yes          | Yes          | Yes          | Yes          |
| Learning Curve         | Steep        | Low          | Low          | Low          | Low          |
| External Dependency    | None         | Git forge    | Git forge    | GitHub.com   | None         |

---

## Decision

**Tekton** is chosen for the following reasons:

1. **CNCF Incubating**: Aligns with the project's CNCF ecosystem philosophy
2. **Kubernetes-native**: Pipelines are CRDs, runs are Pods — fully declarative and GitOps-compatible
3. **ARM64 support**: Full native support with documented multi-arch builds
4. **Scalability**: Scales naturally with Kubernetes, ready for hardware expansion (D17)
5. **Kaniko integration**: Rootless image builds without Docker-in-Docker privilege escalation

The higher resource footprint and learning curve are acceptable trade-offs given the project's focus on learning and CNCF alignment. The planned hardware expansion (Jetson Orin NX, Turing RK1) will provide additional capacity.

---

## Implementation Plan

1. Add Tekton Pipelines to the core-components chart ApplicationSet (remote Helm chart)
2. Create `core-components/tekton/` with values.yaml and manifests
3. Install Tekton Dashboard for pipeline visibility
4. Create initial pipeline: build container image → push to Harbor
5. Configure GitHub webhook for push-triggered builds
6. Document setup in `core-components/tekton/README.md`

---

## References

- [Tekton Documentation](https://tekton.dev/)
- [Tekton Multi-Architecture ARM64 Builds](https://danmanners.com/posts/2022-08-tekton-cicd-multiarch-builds/)
- [Woodpecker CI](https://woodpecker-ci.org/)
- [Woodpecker Homelab Setup](https://pwa.io/articles/installing-woodpecker-in-your-homelab/)
- [Drone CI](https://www.drone.io/)
- [Forgejo Actions](https://forgejo.org/docs/next/admin/actions/)
- D17: Talos migration (hardware expansion context)
