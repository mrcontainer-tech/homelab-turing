# Homelab Kubernetes Cluster

A GitOps-managed Kubernetes homelab I built for learning, experimentation, and running real workloads. Everything is declarative—this Git repository is the single source of truth, with ArgoCD handling continuous deployment.

## Cluster Hardware

| Node | Hardware | Architecture | Role | Notes |
|------|----------|--------------|------|-------|
| node1 | Raspberry Pi CM4 | ARM64 | Control plane | Turing Pi v2.5 slot 1 |
| node2 | Raspberry Pi CM4 | ARM64 | Control plane | Turing Pi v2.5 slot 2 |
| node3 | Raspberry Pi CM4 | ARM64 | Control plane | Turing Pi v2.5 slot 3 |
| node4 | Raspberry Pi CM4 | ARM64 | Worker | Turing Pi v2.5 slot 4,  SSD storage |

The Turing Pi v2.5 hosts four CM4 modules in a compact form factor.

## Repository Structure

```
homelab-turing/
├── core-components/       # Infrastructure layer (MetalLB, Longhorn, cert-manager, etc.)
├── applications/          # Workloads (monitoring, media server)
├── *-application-set.yaml # ArgoCD ApplicationSets (list + git generators)
└── adr/                   # Architecture Decision Records
```

## How It Works

I use ArgoCD **ApplicationSets** to automatically discover and deploy components:

1. **Helm charts** — Multi-source ApplicationSets using a **list generator**. Each entry defines the chart name, remote Helm repo URL, and version. ArgoCD pulls the chart from the upstream repo and merges it with `values.yaml` from this Git repository.
2. **Manifests** — Git generator ApplicationSets scan for `manifests/` directories and deploy raw Kubernetes YAML.
3. Changes pushed to Git trigger automatic sync with prune and self-heal.

### Component Structure

```
component-name/
├── manifests/   # Raw Kubernetes YAML (namespace, CRDs, configs)
├── values.yaml  # Custom Helm values (chart itself comes from remote repo)
└── README.md    # Component documentation
```

Helm charts are **not vendored** — they are referenced from upstream repositories. Only custom value overrides are stored in Git. To update a chart version, change the `version` field in the ApplicationSet.

## Core Components

| Component | Purpose |
|-----------|---------|
| ArgoCD | GitOps controller |
| MetalLB | Bare-metal load balancer |
| Longhorn | Distributed block storage |
| cert-manager | TLS certificate management |
| external-dns | Automatic DNS (Route53) |
| external-secrets | Secrets from external stores |
| dnscrypt-proxy | Encrypted DNS |
| Traefik | Ingress controller |

## Applications

| Application | Purpose |
|-------------|---------|
| kube-prometheus-stack | Monitoring (Prometheus + Grafana) |
| media-server | Jellyfin + Jellyseerr |
| Harbor | Container registry |
| Kyverno | Policy engine |
| Policy Reporter | Policy enforcement reporting |
| Homepage | Homelab landing page |

## Quick Commands

```bash
# ArgoCD
argocd app list
argocd app sync <app-name>

# Check deployments
kubectl get pods -A
kubectl get applications -n argocd
```

## Documentation

- `CLAUDE.md` — AI assistant guide and project conventions
- `VISION.md` — Project philosophy and rationale
- `TURING-INSTALL.md` — Initial cluster setup
- `adr/` — Architecture Decision Records

## Known Issues

**ArgoCD TLS**: Currently disabled via kubectl patch. TODO: Replace with proper Helm values.

```bash
kubectl -n argocd patch deployment argocd-server \
  --type='json' \
  -p='[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--insecure"}]'
```
