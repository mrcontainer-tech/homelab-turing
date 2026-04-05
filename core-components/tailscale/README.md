# Tailscale

Tailscale Kubernetes operator providing secure remote access to homelab services.

## Architecture

Hybrid approach with two exposure mechanisms:

1. **Subnet Router** (Connector CRD) — advertises the MetalLB IP range (`192.168.68.64/28`), giving Tailscale devices full access to all `*.mrcontainer.nl` services
2. **Tailscale Ingress** — exposes priority services with MagicDNS names (`*.tailnet-name.ts.net`) and auto-provisioned TLS

Traefik remains the LAN ingress controller — both coexist.

## Components

| Resource | Purpose |
|----------|---------|
| Helm chart (`tailscale-operator`) | Deploys the operator and CRDs |
| `manifests/namespace.yaml` | Tailscale namespace |
| `manifests/external-secret.yaml` | Syncs OAuth credentials from AWS Secrets Manager |
| `manifests/connector.yaml` | Subnet router for MetalLB IP range |

## Prerequisites

1. **Tailscale account** with an OAuth client created in the admin console
2. **OAuth credentials** stored in AWS Secrets Manager at `homelab/tailscale-operator` with keys `client_id` and `client_secret`
3. **ACL policy** configured in Tailscale admin console with tags (`tag:k8s-operator`, `tag:k8s`, `tag:k8s-subnet`, `tag:k8s-media`) and auto-approver for the subnet route

## Tailscale Ingress Services

Priority services exposed via MagicDNS:

| Service | MagicDNS Name | Backend |
|---------|--------------|---------|
| Homepage | `homepage.tailnet-name.ts.net` | homepage:3000 |
| Jellyfin | `jellyfin.tailnet-name.ts.net` | jellyfin:8096 |
| Jellyseerr | `jellyseerr.tailnet-name.ts.net` | jellyseerr:5055 |
| Grafana | `grafana.tailnet-name.ts.net` | kube-prometheus-stack-chart-grafana:80 |

## Adding a New Service

To expose any service via Tailscale, create an Ingress manifest in the service's `manifests/` directory:

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

Commit to Git and ArgoCD will auto-sync it. The service will appear on MagicDNS with auto-provisioned TLS.

## Related ADRs

- [D22: Tailscale Remote Access](../../adr/D22-tailscale-remote-access.md)
- [D12: CoreDNS Fix](../../adr/) — Tailscale DNS avoidance workaround
- [D21: Cluster Security](../../adr/D21-cluster-security.md) — NetworkPolicies for Tailscale proxy pods
