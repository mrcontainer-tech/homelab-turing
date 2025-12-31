# Knative Serving Functions

This directory contains Knative Serving function definitions and examples for the homelab.

## Structure

```
functions/
├── python-etl-example/      # Python ETL function example
├── go-api-poller/           # Go API polling function
└── manifests/               # Kubernetes manifests for function deployments
```

## Quick Start

### Prerequisites

1. Knative Serving installed in the cluster (see `core-components/knative-serving`)
2. Knative CLI (`kn`) installed locally
3. Docker or compatible build tool

### Development Workflow

#### Option 1: Using kn CLI (Recommended for Quick Deploys)

```bash
# Deploy from a container image
kn service create python-etl-example \
  --image docker.io/username/python-etl-example:latest \
  --port 8080

# Update service
kn service update python-etl-example \
  --image docker.io/username/python-etl-example:v2

# List services
kn service list

# Delete service
kn service delete python-etl-example
```

#### Option 2: Using YAML Manifests (Recommended for GitOps)

```bash
# Apply service manifest
kubectl apply -f python-etl-example/service.yaml

# Check status
kubectl get ksvc

# Get service URL
kubectl get ksvc python-etl-example -o jsonpath='{.status.url}'
```

#### Option 3: Traditional Docker Build + Deploy

```bash
# Navigate to function directory
cd python-etl-example

# Build Docker image
docker build -t docker.io/username/python-etl-example:latest .

# Push to registry
docker push docker.io/username/python-etl-example:latest

# Deploy using kn or kubectl
kn service create python-etl-example \
  --image docker.io/username/python-etl-example:latest
```

### Testing Functions

```bash
# Get service URL
URL=$(kn service describe python-etl-example -o url)

# Invoke with curl
curl -X POST $URL \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {"id": "1", "name": "test", "value": "10"}
    ]
  }'

# Or if using MetalLB LoadBalancer
KOURIER_IP=$(kubectl get svc kourier -n kourier-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
curl -X POST http://$KOURIER_IP \
  -H "Host: python-etl-example.default.example.com" \
  -H "Content-Type: application/json" \
  -d '{"items": [{"id": "1", "name": "test", "value": "10"}]}'
```

## Function Examples

### Python ETL Example
A Flask-based ETL function that demonstrates:
- Reading data from HTTP POST requests
- Transforming the data
- RESTful HTTP interface
- Auto-scaling with scale-to-zero

### Go API Poller
A Go HTTP server that demonstrates:
- Polling external APIs
- Processing JSON responses
- Health check endpoints
- Efficient resource usage

## Development Best Practices

### 1. Container Requirements

Knative services must:
- Listen on port **8080** (or configure `PORT` env var)
- Respond to HTTP health checks (GET /)
- Handle graceful shutdown
- Be stateless

### 2. Image Naming

Use semantic versioning:
```bash
docker build -t registry/function:v1.0.0 .
docker push registry/function:v1.0.0
```

Avoid using `:latest` in production.

### 3. Resource Limits

Always set resource limits:
```yaml
resources:
  limits:
    memory: "512Mi"
    cpu: "1000m"
  requests:
    memory: "256Mi"
    cpu: "500m"
```

### 4. Auto-scaling Configuration

Tune scaling based on your workload:
```yaml
annotations:
  autoscaling.knative.dev/min-scale: "0"    # Scale to zero
  autoscaling.knative.dev/max-scale: "10"   # Max replicas
  autoscaling.knative.dev/target: "100"     # Concurrent requests
```

## Creating New Functions

See the example functions for templates. Each function needs:
- HTTP server listening on port 8080
- Dockerfile
- Knative Service manifest (service.yaml)

## Resources

- **Knative Docs**: https://knative.dev/docs/
- **Knative Samples**: https://knative.dev/docs/samples/
- **CNCF Project**: https://www.cncf.io/projects/knative/
