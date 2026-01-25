# Harbor

Local container registry for development and as a pull-through cache. Images are cached locally and scanned by Trivy.

## Access

- **URL**: https://harbor.mrcontainer.nl
- **Port-forward**: `kubectl port-forward svc/harbor -n harbor 8080:80`

## Registry Endpoints

Configured under **Administration → Registries**:

| Name | Provider | Endpoint URL | Auth |
|------|----------|--------------|------|
| dockerhub | Docker Hub | `https://hub.docker.com` | None (public) |
| ghcr | Docker Registry | `https://ghcr.io` | None (public) |
| quay | Docker Registry | `https://quay.io` | None (public) |
| lscr | Docker Registry | `https://lscr.io` | None (public) |
| k8s | Docker Registry | `https://registry.k8s.io` | None (public) |
| gcr | Docker Registry | `https://gcr.io` | None (public) |
| nvcr | Docker Registry | `https://nvcr.io` | None (public) |

## Proxy Cache Projects

| Project | Registry | Example Pull |
|---------|----------|--------------|
| `dockerhub-proxy` | dockerhub | `harbor.mrcontainer.nl/dockerhub-proxy/library/nginx:latest` |
| `ghcr-proxy` | ghcr | `harbor.mrcontainer.nl/ghcr-proxy/gethomepage/homepage:v1.8.0` |
| `quay-proxy` | quay | `harbor.mrcontainer.nl/quay-proxy/jetstack/cert-manager-controller:v1.14.0` |
| `lscr-proxy` | lscr | `harbor.mrcontainer.nl/lscr-proxy/linuxserver/sonarr:latest` |
| `k8s-proxy` | k8s | `harbor.mrcontainer.nl/k8s-proxy/external-dns/external-dns:v0.14.0` |
| `gcr-proxy` | gcr | `harbor.mrcontainer.nl/gcr-proxy/knative-releases/knative.dev/serving/cmd/webhook:latest` |
| `nvcr-proxy` | nvcr | `harbor.mrcontainer.nl/nvcr-proxy/nvidia/k8s-device-plugin:v0.14.0` |

## Regular Projects

| Project | Purpose |
|---------|---------|
| `library` | Local images built for this homelab |

## Adding a New Proxy Cache

1. **Add Registry Endpoint**: Administration → Registries → + New Endpoint
   - Provider: Docker Registry (or Docker Hub for docker.io)
   - Endpoint URL: `https://<registry-domain>`
   - Test connection before saving

2. **Create Proxy Project**: Projects → + New Project
   - Enable "Proxy Cache"
   - Select the registry endpoint
   - Set access level (Public/Private)