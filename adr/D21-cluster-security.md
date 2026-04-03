# D21: Layered Cluster Security

**Date**: 2026-04-03
**Status**: Proposed
**Decision**: Implement layered cluster security using ValidatingAdmissionPolicy, Kyverno, Tetragon, Pod Security Standards, and NetworkPolicies
**Context**: Kyverno deployed with zero policies, no runtime security, no network segmentation — Talos migration (D16) enables Kubernetes-native admission features

---

## Problem Statement

Despite having security-oriented components deployed, the cluster has significant gaps:

| Area | Current State | Risk |
|------|--------------|------|
| Kyverno (D10) | Deployed (v3.6.2) with zero ClusterPolicies | HIGH — policy engine running idle |
| Policy Reporter | Deployed, dashboards empty | No visibility into policy violations |
| Runtime security | Not deployed | HIGH — no process/syscall/file monitoring |
| NetworkPolicies | None | HIGH — all pod-to-pod traffic unrestricted |
| Pod Security Standards | No namespace labels | HIGH — no restrictions on privileged workloads |
| Admission control | Only default K8s admission | No custom validation or mutation |

The Talos migration (D16) will bring vanilla Kubernetes 1.32+, unlocking:
- **ValidatingAdmissionPolicy** (GA since K8s 1.30) — in-process CEL-based validation, no webhook overhead
- **MutatingAdmissionPolicy** (alpha in K8s 1.32) — in-process CEL-based mutation

Harbor provides Trivy image scanning (D20), and OIDC federation (D19) is replacing static credentials — but nothing protects the runtime or enforces admission policies.

## Goals

- Enforce admission control for all workloads (block privileged containers, restrict image registries, require labels and resource limits)
- Implement runtime threat detection for process execution, network anomalies, and file integrity
- Restrict pod-to-pod network traffic per namespace (default-deny)
- Enforce Pod Security Standards via namespace labels
- Generate PolicyReports so Policy Reporter has useful data
- Stay within ARM64 Raspberry Pi resource constraints (~4GB RAM per node)
- All policies stored in Git, synced by ArgoCD

---

## Options Considered

### Admission Control

#### Option A: ValidatingAdmissionPolicy Only (Kubernetes-native)

Use only the Kubernetes-native ValidatingAdmissionPolicy with CEL expressions.

**Pros**:
- Zero webhook overhead — runs in-process in the API server
- No external dependencies
- CEL expressions are fast and well-tested
- GA since Kubernetes 1.30

**Cons**:
- Cannot generate resources (e.g., auto-create NetworkPolicies)
- Cannot mutate resources (e.g., inject default security contexts)
- Cannot verify container image signatures
- Cannot reference external data sources
- Does not produce PolicyReport resources (Policy Reporter stays empty)

---

#### Option B: Kyverno Only

Use Kyverno for all admission control — validation, mutation, generation, and image verification.

**Pros**:
- Full-featured: validate, mutate, generate, verify images
- Produces PolicyReport resources (feeds Policy Reporter)
- Already deployed (v3.6.2), just needs policies
- Large policy library available
- Supports image verification (cosign/sigstore)

**Cons**:
- All API requests routed through webhook — latency overhead
- Higher resource cost on constrained ARM64 nodes
- Single point of failure if admission controller is unavailable

---

#### Option C: Layered Approach (ValidatingAdmissionPolicy + Kyverno)

Use ValidatingAdmissionPolicy for simple, high-frequency validations. Use Kyverno for complex policies that require generation, image verification, or external data.

**Pros**:
- Simple checks run in-process (no webhook overhead)
- Kyverno only handles complex policies — reduced webhook load
- Best performance-to-capability ratio
- Kyverno produces PolicyReports for Policy Reporter
- Progressive — native policies are the future direction of Kubernetes

**Cons**:
- Two admission systems to manage
- Developers need to understand both CEL and Kyverno policy syntax
- Policy placement decisions (which system handles what)

---

### Runtime Security

#### Option 1: Tetragon

eBPF-based runtime security from the Cilium project (CNCF).

**Pros**:
- Lightest resource footprint (~50-100MB RSS) — critical for CM4 nodes
- eBPF-native — hooks directly into kernel without modules
- ARM64 support
- Enforcement capable — can kill processes, block syscalls, not just detect
- Process execution, network observability, and file integrity monitoring
- Prometheus metrics and Grafana dashboards available
- Helm chart for easy deployment
- TracingPolicy CRDs for declarative, GitOps-friendly configuration

**Cons**:
- Smaller community than Falco
- Younger project
- Requires kernel eBPF support (available on Talos)

---

#### Option 2: Falco

eBPF-based runtime security (CNCF Incubating).

**Pros**:
- Largest community and most mature rule ecosystem
- ARM64 support
- eBPF or kernel module driver options
- Rich rule language for syscall monitoring and anomaly detection
- Extensive documentation and integrations

**Cons**:
- Heavier resource usage (~200-400MB RSS) — significant on 4GB CM4 nodes
- Detection-only — cannot enforce (kill/block), only alert
- More complex configuration
- Falcosidekick required for alerting integrations

---

#### Option 3: KubeArmor

eBPF-based container-aware security enforcement (CNCF Sandbox).

**Pros**:
- eBPF-based, ARM64 support
- Container-aware policy enforcement
- Lightweight

**Cons**:
- CNCF Sandbox — less mature than Tetragon or Falco
- Smaller community and documentation
- Fewer integrations

---

#### Option 4: No Runtime Security

Keep the current state — no runtime monitoring.

**Pros**:
- No additional resource usage

**Cons**:
- No visibility into container-level threats
- No detection of compromised workloads or lateral movement

---

### MutatingAdmissionPolicy

MutatingAdmissionPolicy is alpha in Kubernetes 1.32 and beta in 1.33. While promising (in-process mutation with CEL, no webhook), the API is not yet stable. **Defer to Kyverno mutate rules until MutatingAdmissionPolicy reaches GA** to avoid alpha/beta API churn.

---

## Comparison Matrices

### Admission Control

| Criteria | VAP Only | Kyverno Only | Layered (VAP + Kyverno) |
|----------|----------|--------------|-------------------------|
| Webhook Overhead | None | All requests | Complex requests only |
| Validation | CEL | YAML/JMESPath | Both |
| Resource Generation | No | Yes | Yes (Kyverno) |
| Image Verification | No | Yes (cosign) | Yes (Kyverno) |
| External Data | No | Yes | Yes (Kyverno) |
| PolicyReport Output | No | Yes | Yes (Kyverno) |
| Resource Cost | Minimal | Medium | Low-Medium |
| K8s Version Required | 1.30+ | Any | 1.30+ |

### Runtime Security

| Criteria | Tetragon | Falco | KubeArmor |
|----------|----------|-------|-----------|
| CNCF Status | Graduated (Cilium) | Incubating | Sandbox |
| ARM64 Support | Yes | Yes | Yes |
| Memory Usage | ~50-100MB | ~200-400MB | ~100-200MB |
| eBPF Native | Yes | Yes | Yes |
| Process Monitoring | Yes | Yes | Yes |
| Network Observability | Yes | Limited | Limited |
| File Integrity | Yes | Yes | Yes |
| Enforcement (kill/block) | Yes | No (detect only) | Yes |
| Prometheus Metrics | Yes | Yes | Yes |
| Community Size | Growing | Largest | Small |
| GitOps (CRDs) | TracingPolicy | FalcoRule | KubeArmorPolicy |

---

## Decision

### 1. Layered Admission Control (Option C)

**ValidatingAdmissionPolicy** for simple, high-frequency checks that benefit from zero webhook overhead. **Kyverno** for complex policies requiring generation, image verification, or mutation. This reduces webhook load on resource-constrained CM4 nodes while retaining Kyverno's full capabilities.

### 2. Tetragon for Runtime Security (Option 1)

Tetragon is chosen for its minimal resource footprint (~50-100MB vs Falco's ~200-400MB). On 4GB CM4 nodes, every megabyte matters. Tetragon's enforcement capabilities (kill/block, not just detect) and TracingPolicy CRDs fit the GitOps model. The Cilium/CNCF ecosystem provides long-term confidence.

### 3. Defer MutatingAdmissionPolicy

Use Kyverno mutate rules until MutatingAdmissionPolicy reaches GA. Avoids alpha API instability.

### 4. Pod Security Standards via Namespace Labels

Enforce `restricted` profile on application namespaces, `baseline` on system namespaces. Zero new components — immediate security improvement.

### 5. NetworkPolicies via Kyverno Generate Rules

Kyverno auto-generates a default-deny NetworkPolicy when a namespace is created. Explicit allow rules are added per component in their `manifests/` directories.

---

## Implementation Plan

### Phase 1: Pod Security Standards (immediate — zero new components)

Add labels to namespace manifests:

**Application namespaces** (media-server, harbor, homepage, tokito):
```yaml
metadata:
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/warn: restricted
```

**System namespaces** (argocd, cert-manager, external-dns, external-secrets, kyverno, longhorn-system, metallb-system, kube-prometheus-stack):
```yaml
metadata:
  labels:
    pod-security.kubernetes.io/enforce: baseline
    pod-security.kubernetes.io/warn: restricted
```

### Phase 2: ValidatingAdmissionPolicy (after Talos migration — K8s 1.30+)

Create `core-components/admission-policies/manifests/` with:

| Policy | CEL Expression | Purpose |
|--------|---------------|---------|
| `deny-privileged-containers.yaml` | Block `securityContext.privileged: true` | Prevent privileged escalation |
| `require-harbor-registry.yaml` | Block images not from `harbor.mrcontainer.nl` | Enforce Harbor proxy (D10) |
| `require-resource-limits.yaml` | Require `resources.limits.memory` and `resources.limits.cpu` | Prevent unbounded resource usage |
| `require-labels.yaml` | Require `app.kubernetes.io/name` label | Enforce labeling standards |

### Phase 3: Kyverno Policies

Add ClusterPolicy resources to `applications/kyverno/manifests/`:

| Policy | Type | Purpose |
|--------|------|---------|
| `generate-default-deny-networkpolicy.yaml` | Generate | Auto-create default-deny NetworkPolicy per namespace |
| `verify-image-signatures.yaml` | Verify | cosign/sigstore image verification |
| `mutate-default-security-context.yaml` | Mutate | Inject `runAsNonRoot`, `readOnlyRootFilesystem`, drop capabilities |
| `mutate-default-labels.yaml` | Mutate | Add standard labels if missing |

These policies will generate PolicyReports for Policy Reporter.

### Phase 4: Tetragon Deployment

1. Add to `applications-chart-application-set.yaml`:
   - Chart: `tetragon`, Repo: `https://helm.cilium.io`, Version: latest stable
2. Create `applications/tetragon/values.yaml` with ARM64 resource limits
3. Create `applications/tetragon/manifests/namespace.yaml`
4. Deploy TracingPolicy resources for:
   - Process execution monitoring (detect unexpected binaries)
   - File integrity monitoring (`/etc/shadow`, `/etc/passwd`, sensitive mounts)
   - Network connection monitoring (unexpected outbound connections)
5. Configure Prometheus ServiceMonitor for Tetragon metrics

### Phase 5: NetworkPolicy Allow Rules

After Phase 3 generates default-deny policies, add explicit allow rules per namespace:

| Allow Rule | Purpose |
|------------|---------|
| Ingress from Traefik | Allow external traffic to services |
| DNS (port 53) | Allow CoreDNS resolution |
| Prometheus scraping | Allow metrics collection |
| ArgoCD → API server | Allow GitOps sync |

Store in each component's `manifests/` directory.

---

## Dependencies

| Phase | Depends On |
|-------|-----------|
| Phase 1 (PSS) | None — can implement immediately |
| Phase 2 (VAP) | Talos migration (D16) — requires K8s 1.30+ |
| Phase 3 (Kyverno) | None — Kyverno already deployed |
| Phase 4 (Tetragon) | Talos migration (D16) — eBPF kernel support |
| Phase 5 (NetworkPolicy) | Phase 3 (Kyverno generate rules create default-deny) |

---

## References

- [ValidatingAdmissionPolicy](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/)
- [MutatingAdmissionPolicy](https://kubernetes.io/docs/reference/access-authn-authz/mutating-admission-policy/)
- [CEL in Kubernetes](https://kubernetes.io/docs/reference/using-api/cel/)
- [Kyverno Documentation](https://kyverno.io/docs/)
- [Kyverno Generate Rules](https://kyverno.io/docs/writing-policies/generate/)
- [Kyverno Image Verification](https://kyverno.io/docs/writing-policies/verify-images/)
- [Tetragon Documentation](https://tetragon.io/docs/)
- [Tetragon Helm Chart](https://github.com/cilium/tetragon/tree/main/install/kubernetes/tetragon)
- [Falco Documentation](https://falco.org/docs/)
- [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- D10: Use Kyverno to enforce Harbor repository usage
- D16: Talos migration (enables K8s 1.32+)
- D19: OIDC federation for AWS authentication
- D20: Harbor migration to ARM64
