# D22: Tailscale Remote Access for Homelab Services

**Date**: 2026-04-04
**Status**: Accepted
**Decision**: Deploy the Tailscale Kubernetes operator with a hybrid approach — subnet router for full network access plus Tailscale Ingress for priority services
**Context**: Homelab services are only accessible on the local network — need secure remote access from personal devices anywhere

---

## Problem Statement

The cluster exposes 14+ services as `*.mrcontainer.nl` subdomains via Traefik ingress, but all are only reachable from the local network (192.168.68.x). The current setup has these gaps:

| Area | Current State | Impact |
|------|--------------|--------|
| Remote access | None | Cannot reach services from phone/laptop outside home |
| Tailscale integration | Running on nodes (D12 CoreDNS workaround) | No Kubernetes-native integration, no service exposure |
| Service exposure model | Traefik + MetalLB (LAN only) | All-or-nothing; no per-service remote access control |
| DNS for remote | Only mrcontainer.nl via Route53 | No remote-friendly DNS resolution |

Tailscale is already running on the k3s nodes (evidenced by the CoreDNS custom config forwarding to dnscrypt-proxy instead of `/etc/resolv.conf` to avoid Tailscale's 100.100.100.100 DNS). However, there is no Kubernetes operator, no Tailscale-based ingress, and no structured way to access services remotely.

### Use Cases

- Access homepage and media server (Jellyfin, Jellyseerr) from phone while traveling
- Manage the full *arr stack (Sonarr, Radarr, Prowlarr, Bazarr, qBittorrent) remotely
- Monitor cluster health via Grafana from anywhere
- Future services: financial app, bookmarking, photo archive — all need remote access from day one

## Goals

- Access homelab services securely from personal devices (phone, laptop) anywhere
- Priority services: homepage, Jellyfin, Jellyseerr, Grafana
- Easy to add new services over time (single manifest per service)
- Minimal disruption to existing Traefik + cert-manager + external-dns stack
- GitOps-managed deployment via ArgoCD
- Granular access control per device/service via Tailscale ACLs
- Stay within ARM64 Raspberry Pi resource constraints (~4GB RAM per node)

---

## Options Considered

### Option A: Subnet Router Only (Connector CRD)

Deploy a Tailscale Connector resource advertising the MetalLB IP range `192.168.68.70-80`. All Tailscale devices can reach existing services at their `*.mrcontainer.nl` addresses.

**Pros**:
- Simplest approach — zero changes to existing ingress, certs, or DNS
- All 14+ services immediately accessible remotely via existing names
- Single Connector resource, minimal resource usage
- Works with existing Traefik TLS certs

**Cons**:
- All-or-nothing access — every Tailscale device sees every service
- Relies on client DNS resolving `*.mrcontainer.nl` to MetalLB IPs (requires DNS config on devices or split DNS)
- No per-service access control via Tailscale ACLs (ACLs work on IP:port, not service names)

### Option B: Tailscale Operator with Ingress Only

Use `ingressClassName: tailscale` on new Ingress resources to expose specific services on the tailnet. Each gets a MagicDNS name like `jellyfin.tailnet-name.ts.net` with auto-provisioned TLS.

**Pros**:
- Per-service granular exposure — only chosen services appear on tailnet
- Auto-provisioned TLS certs (Let's Encrypt via Tailscale)
- MagicDNS names work on all Tailscale devices without DNS configuration
- ACL tags on each Ingress for fine-grained access control
- No changes to existing Traefik Ingress resources (coexistence)

**Cons**:
- Each exposed service requires a separate Tailscale Ingress resource
- Services get `*.ts.net` names, not custom domain names
- Users must remember two different hostnames per service (LAN vs. remote)
- Each Ingress spawns a Tailscale proxy pod (resource usage per service)

### Option C: Tailscale Operator as Primary Ingress Controller

Replace Traefik entirely with Tailscale Ingress for all services.

**Pros**:
- Single ingress stack
- Automatic TLS without cert-manager for Tailscale-exposed services

**Cons**:
- Tailscale Ingress only works for tailnet devices — breaks LAN access for non-Tailscale devices
- Loses Traefik features (middleware, rate limiting, dashboard)
- Couples all service access to Tailscale availability
- Cannot use custom domains natively (open feature request)

### Option D: Hybrid Approach (Subnet Router + Tailscale Ingress)

Deploy both: a Connector for subnet routing the MetalLB range (full access when needed) plus Tailscale Ingress for high-priority services (convenient MagicDNS access with auto-TLS). Keep Traefik for LAN access unchanged.

**Pros**:
- Best of both worlds: MagicDNS convenience for daily services, full network access when needed
- Existing Traefik + cert-manager stack unchanged
- Per-service ACL control via Tailscale tags on Ingress resources
- Subnet router as fallback for services not individually exposed
- Progressive — start with a few Ingress resources, expand as needed
- Adding a new service is one manifest file — commit and ArgoCD syncs

**Cons**:
- Two exposure mechanisms to understand and manage
- Subnet router routes need approval in Tailscale admin console (or auto-approver ACL)
- More moving parts than Option A alone

---

## Comparison Matrix

| Criteria | A: Subnet Only | B: Ingress Only | C: Replace Traefik | D: Hybrid |
|----------|---------------|-----------------|--------------------|-----------| 
| Setup complexity | Low | Medium | High | Medium |
| Per-service ACL control | No (IP-based) | Yes (tags) | Yes (tags) | Yes (mixed) |
| MagicDNS convenience | No | Yes | Yes | Yes (priority) |
| Existing stack impact | None | None | Removes Traefik | None |
| Auto-TLS for remote | No (uses existing) | Yes | Yes | Yes (priority) |
| Resource cost | 1 pod | 1 pod per service | 1 pod per service | 1 + N pods |
| Custom domain support | N/A (uses existing) | No (ts.net only) | No (ts.net only) | Mixed |
| Device DNS config needed | Yes | No | No | Partial |
| Fallback/full access | Yes | No | No | Yes |

---

## Decision

**Option D: Hybrid Approach**

1. **Tailscale Kubernetes Operator** deployed via Helm chart as a core component, managed by ArgoCD
2. **Connector subnet router** advertising `192.168.68.64/28` (covers MetalLB pool 70-80) for full network access from trusted devices
3. **Tailscale Ingress** for high-priority daily services: homepage, Jellyfin, Jellyseerr, Grafana — these get MagicDNS names with auto-TLS
4. **Traefik remains** as the LAN ingress controller — zero changes to existing Ingress resources, cert-manager Certificates, or external-dns records
5. **ACL policy** in Tailscale to control which devices can access which services

### Rationale

- The subnet router provides immediate access to all services with zero per-service configuration
- Tailscale Ingress for priority services adds the convenience of MagicDNS names that work on phones without DNS configuration
- The existing `*.mrcontainer.nl` stack continues to serve LAN access
- Resource usage is bounded: 1 operator pod + 1 subnet router pod + N Ingress proxy pods
- Adding new services over time is trivial — one Ingress manifest per service, commit to Git, ArgoCD syncs

---

## Domain Strategy

- **mrcontainer.nl** continues for LAN access (Traefik + cert-manager + external-dns)
- **MagicDNS (`*.tailnet-name.ts.net`)** for Tailscale remote access with auto-provisioned TLS
- **Second custom domain**: Deferred. Tailscale does not natively support custom domains for Ingress TLS certificates (open feature request). If this feature ships, a second domain could be configured for Tailscale HTTPS. Until then, MagicDNS provides the remote access naming.

## DNS Considerations

The existing CoreDNS custom config (`core-components/coredns/manifests/coredns-custom.yaml`) forwards to dnscrypt-proxy to avoid Tailscale's 100.100.100.100. This remains correct and should not be changed — in-cluster DNS must not route through Tailscale's DNS resolver. The Tailscale operator manages its own DNS for tailnet devices independently.

## Integration with Existing Stack

| Component | Impact | Action Required |
|-----------|--------|----------------|
| cert-manager | None | Continues issuing certs for `*.mrcontainer.nl` via Route53 DNS-01 |
| external-dns | None | Continues managing Route53 records for Traefik Ingress |
| MetalLB | None | Continues providing LoadBalancer IPs. Subnet router advertises this range. |
| Traefik | None | Remains the default IngressClass. Tailscale Ingress uses `ingressClassName: tailscale` |
| CoreDNS | None | Existing workaround to avoid 100.100.100.100 remains in place |
| external-secrets | Used | OAuth credentials for Tailscale operator synced from AWS Secrets Manager |

---

## Security: ACL Policy

The Tailscale policy file (managed in the Tailscale admin console) should define tags and access rules:

```jsonc
{
  "tagOwners": {
    "tag:k8s-operator": [],
    "tag:k8s":          ["tag:k8s-operator"],
    "tag:k8s-subnet":   ["tag:k8s-operator"],
    "tag:k8s-media":    ["tag:k8s-operator"],
    "tag:k8s-admin":    ["tag:k8s-operator"]
  },
  "autoApprovers": {
    "routes": {
      "192.168.68.64/28": ["tag:k8s-subnet"]
    }
  },
  "acls": [
    // Personal devices can access media and homepage services
    {"action": "accept", "src": ["autogroup:member"], "dst": ["tag:k8s-media:*"]},
    // Personal devices can access all services via subnet router
    {"action": "accept", "src": ["autogroup:member"], "dst": ["tag:k8s-subnet:*"]},
    // Admin access to cluster management services
    {"action": "accept", "src": ["autogroup:owner"],  "dst": ["tag:k8s-admin:*"]}
  ]
}
```

This allows tagging Tailscale Ingress resources with purpose-specific tags (e.g., `tag:k8s-media` for Jellyfin) and controlling access per device group.

---

## Implementation Plan

### Phase 1: Tailscale Operator Deployment

1. Create `core-components/tailscale/` directory with values.yaml, manifests, and README
2. Add Helm chart entry to `appsets/core-components-chart-application-set.yaml`
3. Store Tailscale OAuth client credentials in AWS Secrets Manager
4. Create ExternalSecret to sync credentials into the tailscale namespace
5. Configure Tailscale ACL policy in admin console with tags and auto-approvers

### Phase 2: Subnet Router

1. Deploy Connector manifest advertising `192.168.68.64/28` (covers MetalLB pool)
2. Auto-approve route via ACL policy
3. Test: install Tailscale on phone, verify access to `https://homepage.mrcontainer.nl` via subnet route

### Phase 3: Tailscale Ingress for Priority Services

1. Create Tailscale Ingress manifests alongside existing Traefik Ingress resources:
   - `applications/homepage/manifests/ingress-tailscale.yaml`
   - `applications/media-server/manifests/jellyfin-ingress-tailscale.yaml`
   - `applications/media-server/manifests/jellyseerr-ingress-tailscale.yaml`
   - `applications/kube-prometheus-stack/manifests/grafana-ingress-tailscale.yaml`
2. Test: access `https://jellyfin.tailnet-name.ts.net` from phone on cellular network

### Phase 4: Expand and Harden

1. Add Tailscale Ingress for additional services as needed (sonarr, radarr, prowlarr, etc.)
2. Refine ACL policies based on usage patterns
3. Add future services (financial app, bookmarking, photo archive) with Tailscale Ingress from day one
4. Monitor resource usage of proxy pods on CM4 nodes
5. Evaluate custom domain support if/when Tailscale ships the feature

### Adding a New Service (for Phase 4 and beyond)

To expose any new service via Tailscale, create a single Ingress manifest:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: <service>-tailscale
  namespace: <namespace>
spec:
  ingressClassName: tailscale
  tls:
    - hosts:
        - <service>
  rules:
    - http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: <service>
                port:
                  number: <port>
```

Commit to Git → ArgoCD syncs → service appears on MagicDNS with auto-TLS.

---

## Dependencies

| Phase | Depends On |
|-------|-----------|
| Phase 1 (Operator) | Tailscale OAuth client created in admin console; AWS Secrets Manager entry |
| Phase 2 (Subnet Router) | Phase 1; ACL policy with auto-approver for routes |
| Phase 3 (Ingress) | Phase 1; HTTPS and MagicDNS enabled on tailnet |
| Phase 4 (Expand) | Phase 3 validated |

### Cross-ADR Dependencies

- **D16 (Talos)**: After Talos migration, node-level Tailscale will be removed and the operator becomes the sole Tailscale integration point
- **D19 (OIDC)**: ExternalSecret for Tailscale OAuth credentials requires AWS Secrets Manager access via the existing ClusterSecretStore
- **D21 (Security)**: NetworkPolicies will need to allow traffic from Tailscale proxy pods to backend services

---

## References

- [Tailscale Kubernetes Operator](https://tailscale.com/kb/1236/kubernetes-operator)
- [Cluster Ingress — Tailscale Docs](https://tailscale.com/kb/1439/kubernetes-operator-cluster-ingress)
- [Connector (Subnet Router) — Tailscale Docs](https://tailscale.com/kb/1441/kubernetes-operator-connector)
- [Tailscale Operator Helm Chart](https://pkgs.tailscale.com/helmcharts)
- [ACL Policy — Tailscale Docs](https://tailscale.com/kb/1192/acl-samples)
- [Device Tags — Tailscale Docs](https://tailscale.com/kb/1068/acl-tags)
- D12: CoreDNS fix (Tailscale DNS avoidance)
- D16: Talos migration
- D19: OIDC federation for AWS authentication
- D21: Layered cluster security (NetworkPolicies)
