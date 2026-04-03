# CLAUDE.md - AI Assistant Guide

This document helps AI assistants (like Claude) understand how to work effectively with this homelab-turing project. To enhance my dev workflow with AI tools I created this document.

## Project Context

This is a **GitOps-managed Kubernetes homelab** running on a Turing Pi v2.5 with 4 Raspberry Pi CM4s. The cluster is managed declaratively through ArgoCD, with all configuration stored in this Git repository as the single source of truth.

**Key Philosophy**: This is an engineering playground for learning and experimentation. It showcases CNCF ecosystem tools and Kubernetes patterns. This helps me with tinkering and trying out new things.

---

## Repository Structure

```
homelab-turing/
├── core-components/           # Infrastructure layer
│   ├── argocd/
│   ├── metallb-system/
│   ├── longhorn-system/
│   ├── cert-manager/
│   ├── external-dns/
│   ├── external-secrets/
│   ├── dnscrypt-proxy/
│   ├── coredns/
│   └── namespaces/
│
├── applications/              # Workload layer
│   ├── kube-prometheus-stack/ # Monitoring
│   ├── media-server/          # Jellyfin, Jellyseerr
│   ├── harbor/                # Container registry
│   ├── kyverno/               # Policy engine
│   ├── policy-reporter/       # Policy enforcement reporting
│   ├── homepage/              # Homelab landing page
│   └── tokito/
│
├── core-components-chart-application-set.yaml    # List generator: remote Helm charts
├── core-components-manifest-application-set.yaml  # Git generator: raw YAML manifests
├── applications-chart-application-set.yaml        # List generator: remote Helm charts
└── applications-manifest-application-set.yaml     # Git generator: raw YAML manifests
```

---

## ArgoCD ApplicationSet Pattern

This project uses **ApplicationSets** to automatically discover and deploy components:

### How It Works

1. **ApplicationSets** deploy components using two patterns:
   - **Helm charts**: List generator with multi-source — chart from remote Helm repo, values from Git
   - **Manifests**: Git generator scanning `core-components/*/manifests/` and `applications/*/manifests/`

2. **For each component**, ArgoCD creates a child Application:
   - Name: `<component>-<type>` (e.g., `metallb-system-chart`)
   - Namespace: `<component>` (e.g., `metallb-system`)
   - Auto-sync enabled with prune and self-heal

3. **Helm charts** are referenced from remote repositories (not vendored) with values in Git

### Directory Conventions

Each component follows this structure:

```
component-name/
├── manifests/          # Raw YAML files (CRDs, configs, namespace)
├── values.yaml         # Helm values override (referenced by multi-source ApplicationSet)
└── README.md           # Component documentation
```

---

## Common Tasks

### Adding a New Core Component

1. **Create directory structure**:
   ```bash
   mkdir -p core-components/<component-name>/manifests
   ```

2. **For Helm charts** — add entry to the appropriate ApplicationSet:
   - Add element to the list generator in `core-components-chart-application-set.yaml`
   - Include: name, chart, repoURL, version
   - Create `core-components/<component-name>/values.yaml` with custom overrides

3. **Add manifests** (if needed):
   - Place CRDs, namespace, or additional YAML in `manifests/`

4. **Review and commit** (user responsibility):
   - User will review changes and commit to Git
   - Suggested commit message: `feat: Add <component-name>`

5. **After Git push, ArgoCD will automatically**:
   - Detect the new entries/directories
   - Create Application(s)
   - Deploy to the cluster

### Adding a New Application

Same process as core components, but use `applications/` directory and `applications-chart-application-set.yaml`.

### Updating a Helm Chart Version

1. **Update the version** in the relevant ApplicationSet file (`*-chart-application-set.yaml`)
2. **Adjust values.yaml** if the new chart version requires changes
3. **User reviews and commits** changes to Git
4. **After push, ArgoCD auto-syncs** (prune + self-heal enabled)

### Testing Before Commit

```bash
# Render Helm chart from remote repo with local values
helm repo add <repo-name> <repo-url>
helm template <release-name> <repo-name>/<chart-name> \
  --version <version> \
  -f core-components/<component>/values.yaml

# Validate manifest YAML
kubectl apply --dry-run=client -f core-components/<component>/manifests/
```

---

## Important Conventions

### Naming

- **Core components**: Use `-system` suffix (e.g., `metallb-system`, `longhorn-system`)
- **Applications**: Use descriptive names (e.g., `media-server`, `kube-prometheus-stack`)
- **Namespaces**: Match the component name

### Helm Charts (Multi-Source)

**Charts are referenced from remote Helm repositories** (not vendored):
- ArgoCD multi-source ApplicationSets pull charts from upstream repos
- Only `values.yaml` overrides are stored in Git
- Chart name, repo URL, and version are defined in the ApplicationSet list generator
- To update a chart version, change `version` in the ApplicationSet

### Namespace Management

- Core components create their own namespaces
- Application namespaces are pre-created in `core-components/namespaces/manifests/`

### Values Files

- `values.yaml` at component root — custom overrides for the remote Helm chart
- Chart defaults come from the upstream Helm repository

---

## Technology Stack Reference

### Core Components

| Component | Purpose | Notes |
|-----------|---------|-------|
| **ArgoCD** | GitOps controller | Currently uses kubectl patch for TLS config (TODO: Helm install) |
| **MetalLB** | Load balancer | Provides LoadBalancer IPs on bare metal |
| **Traefik** | Ingress controller | Deployed separately (not bundled with Talos) |
| **Longhorn** | Block storage | SSD on Node 3, S3 backup |
| **cert-manager** | Certificate management | Route53 DNS-01 challenge |
| **external-dns** | DNS automation | Route53 integration |
| **external-secrets** | Secrets management | Sync from external sources |
| **CoreDNS** | DNS | Custom config to avoid Tailscale DNS |
| **dnscrypt-proxy** | Encrypted DNS | Multi-provider DoH/DNSCrypt |

### Applications

| Application | Purpose |
|-------------|---------|
| **kube-prometheus-stack** | Monitoring (Prometheus + Grafana) |
| **Harbor** | Container registry |
| **Kyverno** | Policy engine |
| **media-server** | Jellyfin + Jellyseerr |
| **Policy Reporter** | Policy enforcement reporting |
| **Homepage** | Homelab landing page |

---

## Known Issues & TODOs

### ArgoCD TLS Configuration

Currently disabled via `kubectl patch`:
```bash
kubectl -n argocd patch deployment argocd-server \
  --type='json' \
  -p='[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--insecure"}]'
```

**TODO**: Replace with proper Helm install and values.yaml for GitOps compliance.

---

## Working with This Project (AI Assistant Guidelines)

### CRITICAL: Git Workflow

**NEVER use `git add`, `git commit`, or `git push` commands.**

The user will handle all Git operations manually. Your role is to:
- Create, modify, or delete files as requested
- Validate changes locally (dry-run, linting, testing)
- Inform the user what changes were made
- Let the user review and commit changes themselves

This ensures the user maintains full control over what enters the Git history.

### CRITICAL: kubectl is READ-ONLY

**Only use `kubectl get`, `kubectl describe`, `kubectl logs`, and other read operations.**

**NEVER use `kubectl apply`, `kubectl delete`, `kubectl patch`, `kubectl edit`, or any command that modifies cluster state.** All cluster changes must go through GitOps (commit to Git → ArgoCD syncs).

---

### Additional Use Cases

The AI assistant is also used for:
- **Research**: Documentation, best practices, examples, and architectural decisions
- **Troubleshooting**: Analyzing errors, debugging cluster issues, investigating sync failures

When troubleshooting, ask for logs/output, suggest diagnostic commands, and consider ARM64/resource constraints.

---

### When Adding/Modifying Components

1. **Always check existing patterns** before creating new structures
2. **Follow the directory conventions** (manifests/ and values.yaml)
3. **Reference Helm charts from remote repos** — add to the ApplicationSet list generator
4. **Test locally** with `helm template` or `kubectl apply --dry-run`
5. **Document changes** in component README.md
6. **Inform the user** of all changes made for their review

### When Troubleshooting

1. **Check ArgoCD UI/CLI** for sync status
2. **Review ApplicationSet** generators for discovery issues
3. **Validate YAML** structure and Kubernetes API versions
4. **Check namespaces** - components deploy to their named namespace
5. **Review recent commits** - GitOps means Git is source of truth

### When Making Recommendations

1. **Prefer GitOps approach** - everything in Git
2. **Align with CNCF ecosystem** - this is a learning environment
3. **Keep it declarative** - avoid imperative kubectl commands
4. **Document thoroughly** - this is a showcase project
5. **Consider the Raspberry Pi constraints** - resource-limited environment

---

## Useful Commands

```bash
# ArgoCD app status
argocd app list
argocd app get <app-name>
argocd app sync <app-name>

# Check ApplicationSet generation
kubectl get appset -n argocd
kubectl get applications -n argocd

# Kubernetes debugging
kubectl get pods -A
kubectl get events -A --sort-by='.lastTimestamp'
kubectl describe <resource> <name> -n <namespace>

# Helm operations (local testing with remote charts)
helm repo add <repo-name> <repo-url>
helm template <name> <repo-name>/<chart> --version <ver> -f values.yaml
```

---

## Architecture Decision Records (ADR)

The `adr/` directory contains architectural decisions. Check there for context on technology choices and patterns.

---

## Additional Documentation

- `VISION.md` - Project philosophy and component rationale
- `README.md` - Quick start and operational guide
- `TURING-INSTALL.md` - Initial cluster setup
- Component-specific: `core-components/<component>/README.md`
