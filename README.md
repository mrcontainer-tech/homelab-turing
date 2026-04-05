# Homelab Turing Pi

My personal Kubernetes playground, running on a Turing Pi v2.5 with four Raspberry Pi CM4 modules. This is where I tinker with CNCF tools, break things on purpose, and run real workloads that I actually use every day.

Everything is managed through GitOps. This repository is the single source of truth, and ArgoCD takes care of keeping the cluster in sync. Push to main, and it just works.

## Hardware

| Node | Hardware | Role | Notes |
|------|----------|------|-------|
| node1 | Raspberry Pi CM4 (ARM64) | Control plane | Turing Pi v2.5 slot 1 |
| node2 | Raspberry Pi CM4 (ARM64) | Control plane | Turing Pi v2.5 slot 2 |
| node3 | Raspberry Pi CM4 (ARM64) | Control plane | Turing Pi v2.5 slot 3 |
| node4 | Raspberry Pi CM4 (ARM64) | Worker | Turing Pi v2.5 slot 4, SSD storage |

Four nodes, one board, fits on my desk. The Turing Pi v2.5 is a mini-ITX board that slots in four compute modules and gives you a real cluster to work with.

## What's Running

### Core Infrastructure

The foundation that makes everything else possible.

| Component | What it does |
|-----------|-------------|
| **ArgoCD** | GitOps controller, syncs this repo to the cluster |
| **MetalLB** | Load balancer for bare metal (no cloud needed) |
| **Longhorn** | Distributed block storage with S3 backups |
| **Traefik** | Ingress controller, handles all HTTP/HTTPS routing |
| **cert-manager** | Automatic TLS certificates via Let's Encrypt + Route53 |
| **external-dns** | Auto-creates DNS records in Route53 |
| **external-secrets** | Pulls secrets from AWS Secrets Manager |
| **CloudNativePG** | PostgreSQL operator for managed databases |
| **Tailscale** | Secure remote access from anywhere |
| **CoreDNS + dnscrypt-proxy** | Custom DNS with encrypted upstream resolvers |
| **Tekton** | Kubernetes-native CI/CD pipelines |

### Applications

The stuff I actually use.

| App | What it does |
|-----|-------------|
| **Homepage** | Dashboard that ties everything together |
| **Jellyfin** | Media streaming (movies, TV shows) |
| **Miniflux** | RSS reader for tech news |
| **Linkding** | Bookmark manager |
| **Mealie** | Recipe manager with meal planning |
| **Harbor** | Container registry and pull-through cache |
| **Kyverno + Policy Reporter** | Policy engine and compliance reporting |
| **Prometheus + Grafana** | Monitoring and dashboards |

## How It Works

ArgoCD uses **ApplicationSets** to automatically discover and deploy everything:

- **Helm charts** use a list generator. Each entry points to a remote chart repo and version. Values come from this repo. Charts are never vendored.
- **Manifests** use a git generator that scans `*/manifests/` directories and applies raw YAML.
- Every sync has prune and self-heal enabled. Drift gets corrected automatically.

### Repository Layout

```
homelab-turing/
├── core-components/           # Infrastructure layer
│   ├── argocd/
│   ├── metallb-system/
│   ├── longhorn-system/
│   ├── cert-manager/
│   ├── external-dns/
│   ├── external-secrets/
│   ├── cloudnativepg/
│   ├── tailscale/
│   ├── tekton-operator/
│   ├── coredns/
│   └── dnscrypt-proxy/
├── applications/              # Workloads
│   ├── homepage/
│   ├── media-server/
│   ├── harbor/
│   ├── kube-prometheus-stack/
│   ├── kyverno/
│   ├── policy-reporter/
│   ├── linkding/
│   ├── mealie/
│   └── miniflux/
├── adr/                       # Architecture Decision Records
└── *-application-set.yaml     # ArgoCD ApplicationSets
```

Each component follows the same pattern:

```
component-name/
├── manifests/   # Raw Kubernetes YAML
├── values.yaml  # Helm value overrides (chart comes from upstream)
└── README.md    # What it does and how it's configured
```

## Network Overview

```
                        ┌─────────────────────────────────────────────┐
                        │              Internet                       │
                        └──────────┬──────────────────┬───────────────┘
                                   │                  │
                            ┌──────▼──────┐    ┌──────▼──────┐
                            │  Route 53   │    │  Tailscale  │
                            │  (DNS)      │    │  DERP relay │
                            └──────┬──────┘    └──────┬──────┘
                                   │                  │
            *.mrcontainer.nl       │                  │  MagicDNS (*.ts.net)
                                   │                  │
                        ┌──────────▼──────────────────▼───────────────┐
                        │            Turing Pi v2.5                   │
                        │                                             │
                        │  ┌─────────────────────────────────────┐    │
                        │  │         MetalLB (L2 mode)           │    │
                        │  │       192.168.68.70 - .80           │    │
                        │  └──────────────┬──────────────────────┘    │
                        │                 │                           │
                        │  ┌──────────────▼──────────────────────┐    │
                        │  │      Traefik (Ingress Controller)   │    │
                        │  │      TLS via cert-manager +         │    │
                        │  │      Let's Encrypt DNS-01           │    │
                        │  └──────────────┬──────────────────────┘    │
                        │                 │                           │
                        │  ┌──────────────▼──────────────────────┐    │
                        │  │           Services                  │    │
                        │  │  Homepage, Jellyfin, Grafana,       │    │
                        │  │  Miniflux, Linkding, Mealie,        │    │
                        │  │  Harbor, ArgoCD, Longhorn, ...      │    │
                        │  └─────────────────────────────────────┘    │
                        │                                             │
                        │  ┌─────────────────────────────────────┐    │
                        │  │         Tailscale Operator          │    │
                        │  │  Subnet router: 192.168.68.64/28    │    │
                        │  │  + Tailscale Ingress per service    │    │
                        │  └─────────────────────────────────────┘    │
                        │                                             │
                        │  ┌─────────────────────────────────────┐    │
                        │  │            DNS Chain                │    │
                        │  │  CoreDNS ──► dnscrypt-proxy         │    │
                        │  │          ──► Cloudflare DoH         │    │
                        │  │          ──► Quad9 DNSCrypt         │    │
                        │  └─────────────────────────────────────┘    │
                        │                                             │
                        │  node1 ─── node2 ─── node3 ─── node4        │
                        │  (CP)      (CP)      (CP)      (Worker)     │
                        └─────────────────────────────────────────────┘
```

Two paths into the cluster:

- **LAN**: `*.mrcontainer.nl` resolves via Route53 to MetalLB IPs. Traefik terminates TLS with auto-renewed Let's Encrypt certs. external-dns keeps DNS records in sync.
- **Remote**: Tailscale provides a subnet router (full LAN access) plus dedicated Tailscale Ingress for priority services with MagicDNS names. Works from anywhere, no port forwarding needed.

In-cluster DNS goes through dnscrypt-proxy for encrypted resolution via Cloudflare DoH and Quad9 DNSCrypt.

## Secrets Management

No secrets in Git. All credentials live in AWS Secrets Manager and get synced into the cluster through ExternalSecrets. The only manual secret is the AWS IAM credentials for the external-secrets operator itself (bootstrapped once via kubectl).

## Bootstrapping

ArgoCD is the one thing that needs manual installation. After that, it manages itself and everything else.

```bash
helm repo add argo https://argoproj.github.io/argo-helm
helm install argocd argo/argo-cd --namespace argocd --create-namespace

# Then apply the four ApplicationSets
kubectl apply -f core-components-chart-application-set.yaml
kubectl apply -f core-components-manifest-application-set.yaml
kubectl apply -f applications-chart-application-set.yaml
kubectl apply -f applications-manifest-application-set.yaml
```

From here, ArgoCD discovers everything in the repo and deploys it.

## Quick Commands

```bash
# ArgoCD
argocd app list
argocd app sync <app-name>

# Cluster health
kubectl get pods -A
kubectl get applications -n argocd
kubectl get appset -n argocd
```

## Documentation

| Doc | What's in it |
|-----|-------------|
| `VISION.md` | Why this project exists and where it's going |
| `CLAUDE.md` | Conventions for working with AI assistants |
| `TURING-INSTALL.md` | Initial cluster setup from scratch |
| `adr/` | All architecture decisions with context and trade-offs |
