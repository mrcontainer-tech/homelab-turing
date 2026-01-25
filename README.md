# Homelab Kubernetes Cluster

A GitOps-managed Kubernetes homelab I built for learning, experimentation, and running real workloads. Everything is declarative—this Git repository is the single source of truth, with ArgoCD handling continuous deployment.

## Cluster Hardware

| Node | Hardware | Architecture | Role | Notes |
|------|----------|--------------|------|-------|
| node1 | Raspberry Pi CM4 | ARM64 | Control plane | Turing Pi v2.5 slot 1 |
| node2 | Raspberry Pi CM4 | ARM64 | Control plane | Turing Pi v2.5 slot 2 |
| node3 | Raspberry Pi CM4 | ARM64 | Control plane | Turing Pi v2.5 slot 3, SSD storage |
| node4 | Raspberry Pi CM4 | ARM64 | Worker | Turing Pi v2.5 slot 4 |
| node5 | Dell XPS (refurbished) | x86_64 | Worker | NVIDIA GTX 1650 Ti GPU |

The Turing Pi v2.5 hosts four CM4 modules in a compact form factor. Node5 extends the cluster with x86_64 compute and GPU capabilities for workloads that need more power or NVIDIA acceleration.

## Repository Structure

```
homelab-turing/
├── core-components/       # Infrastructure layer (MetalLB, Longhorn, cert-manager, etc.)
├── applications/          # Workloads (monitoring, media server, serverless functions)
├── *-application-set.yaml # ArgoCD ApplicationSets for automatic discovery
└── adr/                   # Architecture Decision Records
```

## How It Works

I use ArgoCD **ApplicationSets** to automatically discover and deploy components:

1. ApplicationSets scan for `chart/` and `manifests/` directories
2. Each discovered directory becomes an ArgoCD Application
3. Changes pushed to Git trigger automatic sync with prune and self-heal

### Component Structure

```
component-name/
├── chart/       # Vendored Helm chart
├── manifests/   # Raw Kubernetes YAML
└── README.md    # Component documentation
```

Helm charts are vendored (committed to Git) so I have full visibility into what's deployed.

## Core Components

| Component | Purpose |
|-----------|---------|
| ArgoCD | GitOps controller |
| MetalLB | Bare-metal load balancer |
| Longhorn | Distributed block storage |
| cert-manager | TLS certificate management |
| external-dns | Automatic DNS (Route53) |
| external-secrets | Secrets from external stores |
| Knative Serving | Serverless platform |
| Traefik | Ingress controller |

## Applications

| Application | Purpose |
|-------------|---------|
| kube-prometheus-stack | Monitoring (Prometheus + Grafana) |
| media-server | Jellyfin + Jellyseerr |
| functions | Knative serverless functions |
| Harbor | Container registry |
| Kyverno | Policy engine |

## Quick Commands

```bash
# ArgoCD
argocd app list
argocd app sync <app-name>

# Check deployments
kubectl get pods -A
kubectl get applications -n argocd

# Knative functions
kn service list
```

## Documentation

- `CLAUDE.md` — AI assistant guide and project conventions
- `VISION.md` — Project philosophy and rationale
- `TURING-INSTALL.md` — Initial cluster setup
- `applications/functions/README.md` — Knative serverless guide
- `adr/` — Architecture Decision Records

## Known Issues

**ArgoCD TLS**: Currently disabled via kubectl patch. TODO: Replace with proper Helm values.

```bash
kubectl -n argocd patch deployment argocd-server \
  --type='json' \
  -p='[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--insecure"}]'
```
