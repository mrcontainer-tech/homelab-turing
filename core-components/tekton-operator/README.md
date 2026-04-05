# Tekton Operator

Tekton Operator manages the lifecycle of Tekton components on the cluster. It installs and maintains Tekton Pipelines, Triggers, Chains, and Dashboard based on the `TektonConfig` resource.

## Architecture

```
TektonConfig (CRD) → Tekton Operator → installs/manages:
  ├── Tekton Pipelines (tekton-pipelines namespace)
  ├── Tekton Triggers (event-driven pipeline execution)
  ├── Tekton Chains (supply chain security)
  └── Tekton Dashboard (web UI)
```

## Components

| Resource | Purpose |
|----------|---------|
| `manifests/00-operator-release.yaml` | Tekton Operator v0.78.1 (vendored release YAML) |
| `manifests/01-tekton-config.yaml` | TektonConfig CRD — tells the operator what to install |

## What Gets Installed

The `TektonConfig` uses the `all` profile which installs:

| Component | Purpose | Namespace |
|-----------|---------|-----------|
| **Tekton Pipelines** | Core CI/CD engine — Tasks, Pipelines, TaskRuns, PipelineRuns | tekton-pipelines |
| **Tekton Triggers** | EventListeners for webhook-driven pipeline execution | tekton-pipelines |
| **Tekton Chains** | Supply chain security — signs TaskRun results, generates provenance | tekton-pipelines |
| **Tekton Dashboard** | Web UI for viewing and managing pipelines | tekton-pipelines |

## Configuration

Key settings in `TektonConfig`:

- **`profile: all`** — installs all components including Dashboard
- **`enable-api-fields: beta`** — enables beta API features
- **`set-security-context: true`** — restricts container privileges in TaskRuns
- **Pruner** — auto-cleans completed TaskRuns/PipelineRuns daily (keeps last 5 or 24 hours)

## Upgrading

To upgrade the operator:

1. Download the new release YAML from `https://infra.tekton.dev/tekton-releases/operator/previous/v<VERSION>/release.yaml`
2. Replace `manifests/00-operator-release.yaml`
3. Commit — ArgoCD syncs the update
4. The operator will then upgrade the managed components automatically

## CI Pipeline Flow

Once operational, the target CI workflow is:

```
Git push → Tekton EventListener → PipelineRun → TaskRuns:
  ├── Clone source
  ├── Run tests
  ├── Build image (Kaniko, rootless)
  ├── Push to Harbor
  └── Chains signs the result
```

## Related ADRs

- [D18: CI System](../../adr/D18-ci-system.md) — Decision to use Tekton
- [D20: Harbor on ARM64](../../adr/D20-harbor-on-arm.md) — Container registry for built images
