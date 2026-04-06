# Tekton Pipelines

CI/CD pipeline resources for building container images with Tekton.

## Architecture

```
Git Tag (v*) → GitHub Webhook → Tekton EventListener → PipelineRun:
  ├── git-clone (Hub resolver v0.9)
  ├── test-backend (pytest, parallel) → build-backend (Kaniko)
  └── test-frontend (vitest, parallel) → build-frontend (Kaniko)
```

Builds are triggered by pushing a semver Git tag (e.g., `v1.0.0`). The tag is used as the container image tag.

Images are pushed to the internal Zot registry at `zot-chart.zot.svc:5000`.

## Components

| Resource | Purpose |
|----------|---------|
| `00-namespace.yaml` | Namespace reference (owned by Tekton Operator) |
| `01-rbac.yaml` | `tekton-ci` ServiceAccount, Role, RoleBinding |
| `01-git-credentials-external-secret.yaml` | GitHub SSH deploy key from AWS Secrets Manager |
| `02-dashboard-cert.yaml` | TLS certificate for `tekton.mrcontainer.nl` |
| `03-dashboard-ingress.yaml` | Traefik ingress for Tekton Dashboard |
| `04-dashboard-ingress-tailscale.yaml` | Tailscale ingress for remote access |
| `05-task-pytest.yaml` | Custom Task: run pytest in a Python project |
| `06-task-npm-test.yaml` | Custom Task: run npm tests in a Node.js project |
| `10-pipeline.yaml` | FlowFin CI Pipeline (clone + test + build) |
| `11-trigger-template.yaml` | TriggerTemplate for webhook-driven builds |
| `12-trigger-binding.yaml` | Extracts params from GitHub push payload |
| `13-event-listener.yaml` | Listens for GitHub push events on `main` |
| `14-event-listener-ingress.yaml` | Webhook endpoint at `tekton-webhook.mrcontainer.nl` |

## Pipelines

### flowfin-build

Builds FlowFin frontend and backend container images.

**Parameters:**

| Parameter | Default | Description |
|-----------|---------|-------------|
| `git-url` | `git@github.com:scholtenmartijn/flowfin.git` | Git repository URL |
| `git-revision` | `main` | Branch, tag, or commit SHA |
| `image-tag` | `latest` | Tag for the built images |

**Images produced:**
- `zot-chart.zot.svc:5000/flowfin/backend:<tag>`
- `zot-chart.zot.svc:5000/flowfin/frontend:<tag>`

## Manual PipelineRun

Create a manual build via `kubectl` or the Tekton Dashboard at `https://tekton.mrcontainer.nl`:

```bash
kubectl create -f - <<EOF
apiVersion: tekton.dev/v1
kind: PipelineRun
metadata:
  generateName: flowfin-build-manual-
  namespace: tekton-pipelines
spec:
  pipelineRef:
    name: flowfin-build
  params:
    - name: git-url
      value: "git@github.com:scholtenmartijn/flowfin.git"
    - name: git-revision
      value: "main"
    - name: image-tag
      value: "latest"
  workspaces:
    - name: shared-workspace
      volumeClaimTemplate:
        spec:
          accessModes: ["ReadWriteOnce"]
          storageClassName: longhorn
          resources:
            requests:
              storage: 1Gi
  serviceAccountName: tekton-ci
EOF
```

Monitor progress:

```bash
kubectl get pipelinerun -n tekton-pipelines -w
```

## Triggering a Build

Tag a release in the FlowFin repo:

```bash
git tag v1.0.0
git push origin v1.0.0
```

This triggers the pipeline, which builds and pushes:
- `zot-chart.zot.svc:5000/flowfin/backend:v1.0.0`
- `zot-chart.zot.svc:5000/flowfin/frontend:v1.0.0`

## GitHub Webhook Setup

To enable automatic builds on tag push:

1. Go to **GitHub repo > Settings > Webhooks > Add webhook**
2. Set Payload URL to `https://tekton-webhook.mrcontainer.nl`
3. Set Content type to `application/json`
4. Select "Just the push event"

## Prerequisites

- **Tekton Operator** installed (see `core-components/tekton-operator/`)
- **Zot registry** running (see `applications/zot/`)
- **GitHub SSH deploy key** — public key added to the FlowFin repo, private key stored in AWS Secrets Manager at `homelab/tekton` with property `ssh_private_key`

## Troubleshooting

```bash
# Check pipeline definition
kubectl get pipeline -n tekton-pipelines

# List pipeline runs
kubectl get pipelinerun -n tekton-pipelines

# View logs of a specific run
kubectl logs -n tekton-pipelines -l tekton.dev/pipelineRun=<run-name> --all-containers

# Check EventListener status
kubectl get eventlistener -n tekton-pipelines

# Verify images in Zot
curl -s https://zot.mrcontainer.nl/v2/flowfin/backend/tags/list
curl -s https://zot.mrcontainer.nl/v2/flowfin/frontend/tags/list
```
