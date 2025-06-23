# Homelab Turing Pi (mrcontainer-tech)

This repository bootstraps and manages your Turing Pi Kubernetes cluster via ArgoCD. It defines two main directories:

```
homelab-turing/
├── core-components/   # Cluster-wide infrastructure
└── applications/      # End-user workloads and experiments
```

## Overview

* **Argocd-driven**: This repo is your Git source of truth. ArgoCD watches `main` and reconciles all resources.
* **Cluster**: HA K3s on 4 Raspberry Pi CM4s (Turing Pi v2.5).
* **Storage**: An SSD on Node 3 for Longhorn volumes.

---

## core-components/

Contains ArgoCD `Application` manifests for foundational services:

| Component | Path                      | Description                                      |
| --------- | ------------------------- | ------------------------------------------------ |
| MetalLB   | core-components/metallb/  | LoadBalancer implementation via Helm chart       |
| Longhorn  | core-components/longhorn/ | CSI-based distributed storage, uses the SSD disk |

Each subdir includes an ArgoCD `Application` YAML that points to the upstream Helm chart (via `spec.source.chart`), configures namespaces, and sync policies.

### Example: MetalLB ArgoCD Application

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: metallb
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://helm.lolipop.org
    chart: metallb
    targetRevision: v0.13.10
    helm:
      values: values.yaml
  destination:
    server: https://kubernetes.default.svc
    namespace: metallb-system
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

---

## applications/

Contains ArgoCD `Application` manifests for user workloads. Examples:

| App           | Path                        | Purpose                  |
| ------------- | --------------------------- | ------------------------ |
| homeassistant | applications/homeassistant/ | Home automation          |
| htpc          | applications/htpc/          | Home media server (Plex) |
| experiments   | applications/experiments/   | Playgrounds and testing  |

Each folder has its own `Application` YAML pointing to a chart or Kustomize overlay.

---

## Getting Started

1. **Install ArgoCD** on your cluster if not already:

   ```bash
   kubectl create namespace argocd
   kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
   ```

2. **Configure ArgoCD CLI** (optional):

   ```bash
   argocd login <argocd-server> --username admin --password <password>
   ```

3. **Add this repo** as an ArgoCD Application (or use the UI):

   ```bash
   argocd app create homelab-turing \
     --repo https://github.com/mrcontainer-tech/homelab-turing.git \
     --path . \
     --dest-server https://kubernetes.default.svc \
     --dest-namespace default \
     --sync-policy automated --self-heal
   ```

4. **Sync** the root app:

   ```bash
   argocd app sync homelab-turing
   ```

ArgoCD will recursively apply all `Application` manifests in `core-components/` and `applications/`, deploying your LB, storage, and workloads.

---

## Structure

```
.
├── core-components/
│   ├── metallb/      # ArgoCD Application + Helm values
│   └── longhorn/     # ArgoCD Application + storage config
│
└── applications/
    ├── homeassistant/
    ├── htpc/
    └── experiments/
```

Use `argocd app list` and `argocd app diff <name>` to inspect drift and sync status.