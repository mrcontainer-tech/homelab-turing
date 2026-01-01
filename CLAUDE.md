# CLAUDE.md - AI Assistant Guide

This document helps AI assistants (like Claude) understand how to work effectively with this homelab-turing project.

## Project Context

This is a **GitOps-managed Kubernetes homelab** running on a Turing Pi v2.5 with 4 Raspberry Pi CM4s. The cluster is managed declaratively through ArgoCD, with all configuration stored in this Git repository as the single source of truth.

**Key Philosophy**: This is an over-engineered playground for learning and experimentation. It showcases CNCF ecosystem tools and Kubernetes patterns.

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
│   ├── knative-serving/
│   ├── coredns/
│   └── namespaces/
│
├── applications/              # Workload layer
│   ├── kube-prometheus-stack/ # Monitoring
│   ├── media-server/          # Jellyfin, Jellyseerr
│   ├── functions/             # Knative serverless functions
│   ├── harbor/                # Container registry
│   ├── kyverno/               # Policy engine
│   ├── mariadb/
│   └── tokito/
│
├── core-components-chart-application-set.yaml
├── core-components-manifest-application-set.yaml
├── applications-chart-application-set.yaml
└── applications-manifest-application-set.yaml
```

---

## ArgoCD ApplicationSet Pattern

This project uses **ApplicationSets** to automatically discover and deploy components:

### How It Works

1. **ApplicationSets** scan the repository for directories matching patterns:
   - `core-components/*/chart` - Helm charts
   - `core-components/*/manifest` - Raw YAML manifests
   - `applications/*/chart` - Application Helm charts
   - `applications/*/manifests` - Application manifests

2. **For each discovered directory**, ArgoCD creates a child Application:
   - Name: `<component>-<type>` (e.g., `metallb-system-chart`)
   - Namespace: `<component>` (e.g., `metallb-system`)
   - Auto-sync enabled with prune and self-heal

3. **Helm charts** are vendored (committed to Git) for full visibility

### Directory Conventions

Each component follows this structure:

```
component-name/
├── chart/              # Vendored Helm chart (Chart.yaml, templates/, values.yaml)
├── manifests/          # Raw YAML files (CRDs, configs)
├── values.yaml         # (Optional) Additional values override
└── README.md           # Component documentation
```

---

## Common Tasks

### Adding a New Core Component

1. **Create directory structure**:
   ```bash
   mkdir -p core-components/<component-name>/{chart,manifests}
   ```

2. **Vendor the Helm chart** (if using Helm):
   ```bash
   helm repo add <repo-name> <repo-url>
   helm repo update
   helm pull <repo-name>/<chart-name> \
     --version <version> \
     --untar \
     --untardir core-components/<component-name>/chart
   ```

3. **Customize values** (optional):
   - Edit `chart/values.yaml` or create `values.yaml` alongside

4. **Add manifests** (if needed):
   - Place CRDs or additional YAML in `manifests/`

5. **Review and commit** (user responsibility):
   - User will review changes and commit to Git
   - Suggested commit message: `feat: Add <component-name>`

6. **After Git push, ArgoCD will automatically**:
   - Detect the new directories
   - Create Application(s)
   - Deploy to the cluster

### Adding a New Application

Same process as core components, but use `applications/` directory:

```bash
mkdir -p applications/<app-name>/{chart,manifests}
# ... follow steps above
```

### Updating an Existing Component

1. **Make changes** to chart files or manifests
2. **User reviews and commits** changes to Git
3. **After push, ArgoCD auto-syncs** (prune + self-heal enabled)

### Testing Before Commit

```bash
# Render Helm chart locally
helm template <release-name> core-components/<component>/chart \
  -f core-components/<component>/chart/values.yaml

# Validate YAML
kubectl apply --dry-run=client -f core-components/<component>/manifests/
```

---

## Important Conventions

### Naming

- **Core components**: Use `-system` suffix (e.g., `metallb-system`, `longhorn-system`)
- **Applications**: Use descriptive names (e.g., `media-server`, `kube-prometheus-stack`)
- **Namespaces**: Match the component name

### Helm Chart Vendoring

**Always vendor charts** (commit them to Git):
- Provides full visibility into what's deployed
- Enables Git-based review of template changes
- Avoids external dependencies during sync

### Namespace Management

- Core components create their own namespaces
- Application namespaces are pre-created in `core-components/namespaces/manifests/`

### Values Files

- Primary values: `chart/values.yaml` (within vendored chart)
- Overrides: `values.yaml` at component root (optional)
- Environment-specific: Consider separate values files if needed

---

## Technology Stack Reference

### Core Components

| Component | Purpose | Notes |
|-----------|---------|-------|
| **ArgoCD** | GitOps controller | Currently uses kubectl patch for TLS config (TODO: Helm install) |
| **MetalLB** | Load balancer | Provides LoadBalancer IPs on bare metal |
| **Traefik** | Ingress controller | Default with K3s |
| **Longhorn** | Block storage | SSD on Node 3, S3 backup |
| **cert-manager** | Certificate management | Route53 DNS-01 challenge |
| **external-dns** | DNS automation | Route53 integration |
| **external-secrets** | Secrets management | Sync from external sources |
| **CoreDNS** | DNS | Custom config to avoid Tailscale DNS |
| **Knative Serving** | Serverless platform | CNCF graduated, scale-to-zero functions |

### Applications

| Application | Purpose |
|-------------|---------|
| **kube-prometheus-stack** | Monitoring (Prometheus + Grafana) |
| **Harbor** | Container registry |
| **Kyverno** | Policy engine |
| **media-server** | Jellyfin + Jellyseerr |
| **functions** | Knative serverless functions (Python ETL, Go API poller) |
| **MariaDB** | Database |

---

## Knative Serverless Functions

Located in `applications/functions/`, this project uses Knative Serving for serverless workloads:

- **Auto-scaling**: Scale to zero when idle
- **Revision management**: Automatic versioning
- **Traffic splitting**: Blue/green and canary deployments
- **Examples**: Python ETL, Go API poller

**Quick reference**:
```bash
# Deploy function
kubectl apply -f applications/functions/<function-name>/service.yaml

# Test (get URL from service)
SERVICE_URL=$(kubectl get ksvc <function-name> -o jsonpath='{.status.url}')
curl -X POST $SERVICE_URL -H "Content-Type: application/json" -d '{"test": "data"}'
```

See `applications/functions/QUICKSTART.md` for details.

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

---

### Additional Use Cases

The AI assistant is also used for:
- **Research**: Documentation, best practices, examples, and architectural decisions
- **Troubleshooting**: Analyzing errors, debugging cluster issues, investigating sync failures

When troubleshooting, ask for logs/output, suggest diagnostic commands, and consider ARM64/resource constraints.

---

### When Adding/Modifying Components

1. **Always check existing patterns** before creating new structures
2. **Follow the directory conventions** (chart/ and manifests/)
3. **Vendor Helm charts** - never use remote chart references
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

# Knative functions
kn service list
kn revision list

# Kubernetes debugging
kubectl get pods -A
kubectl get events -A --sort-by='.lastTimestamp'
kubectl describe <resource> <name> -n <namespace>

# Helm operations (local testing)
helm template <name> <chart-path> -f values.yaml
helm lint <chart-path>
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
- Functions: `applications/functions/QUICKSTART.md`
