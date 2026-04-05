# D23: Security Hardening — Credential Hygiene, Workload Security, and Access Reduction

**Date**: 2026-04-05
**Status**: Proposed
**Decision**: Remediate existing security gaps across credentials, workload configurations, RBAC, and service exposure
**Context**: Repository audit revealed hardcoded credentials in git, missing security contexts, overly broad RBAC, and admin services exposed without authentication

---

## Problem Statement

A security audit of the repository revealed concrete vulnerabilities in existing workloads. While D21 establishes the _framework_ for future enforcement (admission policies, Kyverno, Tetragon, NetworkPolicies), this ADR addresses the _remediation_ of existing gaps that can be fixed immediately without new tooling.

### Audit Findings

| Area | Finding | Severity | Files |
|------|---------|----------|-------|
| Hardcoded credentials | Plaintext secrets committed to git | Critical | `media-secret.yaml`, `harbor/values.yaml` |
| Security contexts | No securityContext on any media or homepage deployment | High | 7 deployment manifests |
| RBAC | Homepage has cluster-wide read access | High | `homepage/manifests/crb.yaml` |
| Service exposure | Admin services (ArgoCD, Prometheus, Grafana, Harbor) publicly accessible without auth | High | Multiple ingress files |
| Resource limits | Missing on core Helm components | Medium | cert-manager, external-dns, external-secrets values |
| Ingress consistency | Mixed deprecated annotations and ingressClassName | Medium | Multiple ingress files |
| Label conventions | Inconsistent across components | Medium | Media server deployments |
| Image pinning | Tags used without SHA digests | Medium | All deployments |
| Cleanup | Test ExternalSecret, commented-out files | Low | `test-external-secret.yaml`, `nodes-scheduling.yaml` |

---

## Goals

- Remove all hardcoded credentials from git-tracked files
- Add security contexts to all workload deployments
- Reduce RBAC scope to least privilege
- Move admin/infrastructure services behind Tailscale-only access
- Set resource limits on all Helm-managed components
- Standardize ingress and label conventions
- All changes stored in Git, synced by ArgoCD

---

## Decisions

### 1. Credential Hygiene (Critical)

**Migrate remaining hardcoded secrets to ExternalSecrets.**

| Secret | Current Location | Action |
|--------|-----------------|--------|
| `media-admin` (qbit/jellyfin creds) | `media-server/manifests/media-secret.yaml` | Migrate to ExternalSecret → `homelab/media-admin` |
| Harbor database password | `harbor/values.yaml` line 50 | Move to ExternalSecret → `homelab/harbor-database` |

The pattern is established: ExternalSecret → ClusterSecretStore (`aws-secrets-manager`) → AWS Secrets Manager. Both secrets follow the same approach used for `homepage-api-keys`, `arr-apps-apikeys`, `protonvpn-creds`, `certmanager`, and `external-dns`.

After migration:
- Remove `media-secret.yaml` from git tracking (`git rm --cached`)
- Replace Harbor password in `values.yaml` with a reference to the externally managed secret

### 2. Workload Security Contexts (High)

**Add securityContext to all deployments that currently lack them.**

Target deployments:
- `applications/media-server/manifests/sonarr-deployment.yaml`
- `applications/media-server/manifests/radarr-deployment.yaml`
- `applications/media-server/manifests/prowlarr-deployment.yaml`
- `applications/media-server/manifests/bazarr-deployment.yaml`
- `applications/media-server/manifests/jellyfin-deployment.yaml`
- `applications/media-server/manifests/jellyseerr-deployment.yaml`
- `applications/homepage/manifests/deployment.yaml`

Standard security context for application workloads:
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  runAsGroup: 1000
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
```

**Note**: Some containers (e.g., LinuxServer.io images) may require root. Test each deployment individually and document exceptions. The qBittorrent/Gluetun deployment requires `NET_ADMIN` capability for VPN — this is a documented exception.

### 3. RBAC Scope Reduction (High)

**Reduce homepage ClusterRoleBinding to minimum required permissions.**

Current state: ClusterRole grants read access to pods, nodes, namespaces, ingress (all kinds), metrics across the entire cluster.

Homepage needs cluster-wide access for its Kubernetes widget (it discovers services across namespaces). However:
- Remove `metrics.k8s.io` access unless actively used by widgets
- Remove `gateway.networking.k8s.io` access (not using Gateway API)
- Document why cluster-wide access is required (service discovery dashboard)

### 4. Admin Service Access Reduction (High)

**Move infrastructure/admin services to Tailscale-only ingress, remove public Traefik ingress.**

| Service | Current Access | Target Access |
|---------|---------------|---------------|
| ArgoCD | `argocd.mrcontainer.nl` (public) | Tailscale Ingress only |
| Prometheus | `prometheus.mrcontainer.nl` (public) | Tailscale Ingress only |
| Grafana | `grafana.mrcontainer.nl` (public) | Both (Tailscale primary, Traefik LAN) |
| Harbor | `harbor.mrcontainer.nl` (public) | Both (needs LAN access for in-cluster pulls) |
| Longhorn | `longhorn.mrcontainer.nl` (public) | Tailscale Ingress only |
| Policy Reporter | `policy-reporter.mrcontainer.nl` (public) | Tailscale Ingress only |

Services that should remain on both Traefik (LAN) and Tailscale:
- **Homepage** — landing page, low risk
- **Jellyfin, Jellyseerr** — media consumption from LAN devices (TV, etc.)
- **Harbor** — needs LAN access for cluster image pulls
- **Grafana** — useful on LAN dashboards but primarily remote access

For ArgoCD, Prometheus, Longhorn, and Policy Reporter: remove the Traefik Ingress and Certificate resources, leaving only Tailscale Ingress for remote access and direct ClusterIP for in-cluster communication.

This directly builds on D22 (Tailscale remote access).

### 5. Resource Limits on Core Components (Medium)

**Add explicit resource requests and limits to all Helm-managed components.**

| Component | Values File | Suggested Limits |
|-----------|------------|-----------------|
| cert-manager | `core-components/cert-manager/values.yaml` | 100m/128Mi → 200m/256Mi |
| external-dns | `core-components/external-dns/values.yaml` | 50m/64Mi → 100m/128Mi |
| external-secrets | `core-components/external-secrets/values.yaml` | 100m/128Mi → 200m/256Mi |

These are conservative limits for ARM64 CM4 nodes (~4GB RAM each). Monitor with Prometheus after applying and adjust.

### 6. Ingress Standardization (Medium)

**Migrate all ingress resources to use `ingressClassName` field, remove deprecated annotations.**

Current inconsistencies:
- Some use `kubernetes.io/ingress.class: traefik` annotation (deprecated since K8s 1.22)
- Some use `spec.ingressClassName: traefik` (correct)
- Longhorn uses both simultaneously

Standard pattern going forward:
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: <service>
  namespace: <namespace>
spec:
  ingressClassName: traefik  # or tailscale
  tls:
    - hosts:
        - <host>
      secretName: <cert>
  rules:
    - host: <host>
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: <service>
                port:
                  number: <port>
```

No Traefik-specific annotations (`traefik.ingress.kubernetes.io/*`). No deprecated `kubernetes.io/ingress.class` annotation.

### 7. Label Standardization (Medium)

**Adopt `app.kubernetes.io/*` labels across all workloads.**

Current state: media apps use `app: sonarr`, homepage uses `app.kubernetes.io/name`, dnscrypt-proxy uses `app.kubernetes.io/name`.

Standard labels for all deployments and services:
```yaml
labels:
  app.kubernetes.io/name: <component>
  app.kubernetes.io/part-of: <group>      # e.g., media-server
  app.kubernetes.io/managed-by: argocd
```

This requires updating both deployment selectors and service selectors simultaneously to avoid orphaned ReplicaSets.

### 8. Cleanup (Low)

| Item | Action |
|------|--------|
| `core-components/external-secrets/manifests/test-external-secret.yaml` | Remove — test artifact |
| `core-components/longhorn-system/manifests/nodes-scheduling.yaml` | Remove — entirely commented out |

---

## Implementation Plan

### Phase 1: Credentials (immediate)

1. Create `homelab/media-admin` in AWS Secrets Manager
2. Replace `media-secret.yaml` with ExternalSecret manifest
3. Move Harbor database password to `homelab/harbor-database` in AWS SM
4. Create ExternalSecret for Harbor or use Helm existingSecret pattern
5. Remove `media-secret.yaml` from git tracking

### Phase 2: Access Reduction (after D22 Tailscale is operational)

1. Create Tailscale Ingress for ArgoCD, Prometheus, Longhorn, Policy Reporter
2. Remove public Traefik Ingress and Certificate resources for those services
3. Verify access via Tailscale before removing Traefik ingress

### Phase 3: Workload Hardening

1. Add security contexts to all 7 deployments
2. Test each deployment — document exceptions (e.g., Gluetun needs NET_ADMIN)
3. Add resource limits to cert-manager, external-dns, external-secrets values
4. Reduce homepage RBAC scope

### Phase 4: Standardization

1. Migrate all ingress resources to `ingressClassName` pattern
2. Standardize labels across all deployments and services
3. Clean up test files and commented-out code

---

## Dependencies

| Phase | Depends On |
|-------|-----------|
| Phase 1 (Credentials) | None — ExternalSecrets infrastructure is operational |
| Phase 2 (Access Reduction) | D22 (Tailscale) operational and validated |
| Phase 3 (Workload Hardening) | None — can be done independently |
| Phase 4 (Standardization) | None — can be done independently |

### Cross-ADR Dependencies

- **D21 (Cluster Security)**: Pod Security Standards (Phase 1 of D21) will enforce security context requirements — D23 Phase 3 ensures existing workloads are compliant before D21 enforcement begins
- **D22 (Tailscale)**: Phase 2 depends on Tailscale Ingress being operational for admin service migration
- **D14 (Manual Credentials)**: Fully superseded once Phase 1 is complete

---

## References

- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/overview/)
- [Pod Security Context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)
- [Kubernetes Recommended Labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/)
- [Ingress Class](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class)
- D14: Manual credentials (superseded)
- D21: Layered cluster security (framework)
- D22: Tailscale remote access (enables Phase 2)
