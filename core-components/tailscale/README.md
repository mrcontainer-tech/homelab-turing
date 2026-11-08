# Tailscale

Tailscale Kubernetes operator providing secure remote access to homelab services.

## Architecture

Hybrid approach with two exposure mechanisms:

1. **Subnet Router** (Connector CRD) — advertises the LAN (`192.168.68.0/24`): router `.1`, nodes `.101-.104` and the MetalLB pool, so Tailscale devices reach all `*.mrcontainer.nl` services and the router UI
2. **Tailscale Ingress** — exposes priority services with MagicDNS names (`*.tailnet-name.ts.net`) and auto-provisioned TLS

Traefik remains the LAN ingress controller — both coexist.

## Components

| Resource | Purpose |
|----------|---------|
| Helm chart (`tailscale-operator`) | Deploys the operator and CRDs |
| `manifests/namespace.yaml` | Tailscale namespace |
| `manifests/external-secret.yaml` | Syncs OAuth credentials from AWS Secrets Manager |
| `manifests/connector.yaml` | Subnet router for the LAN (`192.168.68.0/24`) |
| `manifests/crd.yaml` | ClusterRoleBinding: `tailnet-admins` group → `cluster-admin` |

## Prerequisites

1. **Tailscale account** with an OAuth client created in the admin console
2. **OAuth credentials** stored in AWS Secrets Manager at `homelab/tailscale-operator` with keys `client_id` and `client_secret`
3. **ACL policy** configured in Tailscale admin console with tags (`tag:k8s-operator`, `tag:k8s`, `tag:k8s-subnet`, `tag:k8s-media`) and auto-approver for `192.168.68.0/24` on `tag:k8s-subnet`
4. **ACL destination rule** for the LAN, e.g. `dst: ["192.168.68.0/24:*"]`. Subnet traffic is matched on the destination IP, not on the router's tag
5. **Client side**: `tailscale set --accept-routes` on each device that should use the subnet route

## Pod Security

The namespace enforces `privileged`. The Connector pod runs `privileged: true` (sysctls, NET_ADMIN) and is rejected under `baseline`; the operator then retries forever while the Connector still reports Ready.
`warn: restricted` stays on so drift in the Ingress proxy pods is still visible.

## Tailscale Ingress Services

Priority services exposed via MagicDNS:

| Service | MagicDNS Name | Backend |
|---------|--------------|---------|
| Homepage | `homepage.tailnet-name.ts.net` | homepage:3000 |
| Jellyfin | `jellyfin.tailnet-name.ts.net` | jellyfin:8096 |
| Jellyseerr | `jellyseerr.tailnet-name.ts.net` | jellyseerr:5055 |
| Grafana | `grafana.tailnet-name.ts.net` | kube-prometheus-stack-chart-grafana:80 |

## kubectl Access via API Server Proxy

The operator exposes the Kubernetes API on the tailnet as `tailscale-operator.<tailnet>.ts.net` in **auth mode** (`apiServerProxyConfig.mode: "true"`). Callers are authenticated by Tailscale identity and impersonated via a tailnet grant — no cluster credentials needed on the client.

```bash
tailscale configure kubeconfig tailscale-operator
kubectl auth whoami   # shows your tailnet login
```

Requires an ACL rule to `tag:k8s-operator` and a `tailscale.com/cap/kubernetes` grant (see [D22](../../adr/D22-tailscale-remote-access.md#kubernetes-api-access-kubectl-over-tailscale)). Use the MagicDNS name, not the IP — the TLS cert is issued for the `.ts.net` hostname.

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
