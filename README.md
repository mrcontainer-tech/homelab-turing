# Homelab Turing Pi (mrcontainer-tech)

This repository bootstraps and manages your Turing Pi Kubernetes cluster via Argo CD. It defines two main directories:

```
homelab-turing/
├── core-components/   # Cluster-wide infra (charts + manifests)
└── applications/      # End-user workloads and experiments
```

## Overview

* **GitOps-driven**: This repo is your single source of truth.
  Argo CD watches `main` and reconciles everything under `core-components/` and `applications/`.
* **Cluster**: HA K3s on 4× Raspberry Pi CM4s (Turing Pi v2.5).
* **Storage**: SSD on Node 3, managed by Longhorn.

---

## core-components/ (ApplicationSet)

All foundational services (MetalLB CRDs & chart, Longhorn chart) are managed by a single Argo CD **ApplicationSet**. We vendor upstream Helm charts into `chart/` folders and raw manifests (e.g. CRDs) into `manifest/` folders. The ApplicationSet will:

1. **Scan** `core-components/*/{chart,manifest}` directories
2. **Generate** one child Application per sub-folder
3. **Render** Helm charts (with `Chart.yaml` + `templates/`) and plain YAML

### Directory structure

```
core-components/
├── metallb-system/
│   ├── chart/     # vendored MetalLB Helm chart (Chart.yaml, templates/, values.yaml)
│   └── manifest/  # raw YAML CRDs or bootstrapping manifests
└── longhorn-system/
    ├── chart/     # vendored Longhorn Helm chart
    └── manifest/  # (optional) raw manifests if needed
```

### ApplicationSet example

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: core-components
  namespace: argocd
spec:
  # enable Go templating to turn .path maps into names & namespaces
  goTemplate: true
  goTemplateOptions: ["missingkey=error"]

  generators:
  - git:
      repoURL: https://github.com/mrcontainer-tech/homelab-turing.git
      revision: main
      directories:
        - path: core-components/*/chart
        - path: core-components/*/manifest

  template:
    metadata:
      # example: path="core-components/metallb-system/chart"
      # dir .path.path → "core-components/metallb-system"
      # base(...) → "metallb-system"
      # base .path.path → "chart" or "manifest"
      name: '{{ base (dir .path.path) }}-{{ base .path.path }}'
    spec:
      project: default
      source:
        repoURL: https://github.com/mrcontainer-tech/homelab-turing.git
        targetRevision: main
        path: '{{ .path.path }}'
        # if this is a chart directory, Argo CD auto-detects Chart.yaml + templates/
        # you can optionally be explicit:
        # chart: '{{ base (dir .path.path) }}'
        # helm:
        #   releaseName: '{{ base (dir .path.path) }}'
        #   valueFiles:
        #     - values.yaml
      destination:
        server: https://kubernetes.default.svc
        namespace: '{{ base (dir .path.path) }}'
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
```

* **Helm charts**: Any directory under `chart/` containing a `Chart.yaml` and `templates/` is rendered as a Helm chart.
* **Raw manifests**: Directories under `manifest/` are applied as plain YAML.
* **Naming & namespaces**:

  * App name → `<component>-<type>` (e.g. `metallb-system-chart`, `metallb-system-manifest`)
  * Namespace → `<component>` (e.g. `metallb-system`, `longhorn-system`)

---

## applications/

Contains Argo CD **Application** YAMLs (or you can plug into another ApplicationSet) for all your end-user workloads:

```
applications/
├── homeassistant/      # Home automation (Helm chart or Kustomize overlay)
├── htpc/               # Plex / media server
└── experiments/        # Playgrounds and tests
```

Each folder has its own `Application` manifest pointing to a chart or Kustomize directory, with its own `syncPolicy`.

---

## Getting Started

1. **Install Argo CD** (if not already):

   ```bash
   kubectl create namespace argocd
   kubectl apply -n argocd \
     -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
   ```
2. **Connect your GitHub App** (see `/docs/github-app.md` for setup).
3. **Apply the ApplicationSet** for core-components:

   ```bash
   kubectl apply -f core-components/appset-core-components.yaml
   ```
4. **Create root Application** (or use Argo CD UI) to kick off `applications/`:

   ```bash
   argocd app create homelab-turing \
     --repo https://github.com/mrcontainer-tech/homelab-turing.git \
     --path applications \
     --dest-server https://kubernetes.default.svc \
     --dest-namespace default \
     --sync-policy automated --self-heal
   ```
5. **Sync**:

   ```bash
   argocd app sync core-components
   argocd app sync homelab-turing
   ```

---

## Adding New Helm Charts

To add a new Helm chart to **core‑components** or **applications** and ensure you can review every rendered resource:

1. **Choose a component name**: decide if it belongs under `core-components` (cluster infra) or `applications` (workloads).
2. **Create the target directory**:

   * For core-components: `core-components/<component>-system/chart`
   * For applications:  `applications/<component>/chart`
3. **Add and update the Helm repo**:

   ```bash
   helm repo add <repo‑name> <repo‑url>
   helm repo update
   ```
4. **Pull and unpack the chart**:

   ```bash
   helm pull <repo‑name>/<chart‑name> \
     --version <chart‑version> \
     --untar \
     --untardir <path‑to‑chart>
   ```

   This populates the `chart/` directory with `Chart.yaml`, `templates/`, `values.yaml`, etc.
5. **(Optional) Customize values**:

   * Edit `chart/values.yaml` or add additional `values-*.yaml` files alongside.
6. **Commit & push** your changes:

   ```bash
   git add <new chart directory>
   git commit -m "vendor <chart-name> v<version> chart for <component>"
   git push
   ```

On the next Argo CD sync, the full chart directory is treated as plain YAML so every rendered manifest is applied exactly as you have in Git—letting you review diffs directly in Git and in Argo CD.

## Tips & Tricks

* **Sync waves**: use `argocd.argoproj.io/sync-wave` annotations to enforce ordering (e.g. CRDs before charts).
* **Chart overrides**: add `helm.parameters` or `valueFiles` under the `source.helm:` block to customize installs.
* **Drift detection**: use `argocd app diff <app-name>` to see live vs. desired state.
