# Knative Quick Start Guide

Get started with Knative Serving serverless functions in your homelab.

## Prerequisites

- ArgoCD running in your cluster
- Docker installed locally
- kubectl configured
- MetalLB providing LoadBalancer IPs

## Step 1: Deploy Knative (Automatic via ArgoCD)

Knative is automatically deployed via ArgoCD ApplicationSets when you commit the code.

```bash
# Verify ArgoCD picked up Knative
kubectl get applications -n argocd | grep knative

# Check Knative components
kubectl get pods -n knative-serving
kubectl get pods -n kourier-system

# Wait for everything to be ready
kubectl wait --for=condition=Ready pods -n knative-serving --all --timeout=300s
kubectl wait --for=condition=Ready pods -n kourier-system --all --timeout=300s
```

## Step 2: Install Knative CLI

```bash
# macOS
brew install knative/client/kn

# Linux
wget https://github.com/knative/client/releases/download/knative-v1.15.0/kn-linux-amd64
chmod +x kn-linux-amd64
sudo mv kn-linux-amd64 /usr/local/bin/kn

# Verify
kn version
```

## Step 3: Deploy Example Functions

```bash
cd applications/functions/python-etl-example

# Set your Docker registry
export DOCKER_REGISTRY=docker.io/yourusername

# Build and push image
docker build -t $DOCKER_REGISTRY/python-etl-example:v1.0.0 .
docker push $DOCKER_REGISTRY/python-etl-example:v1.0.0

# Update service.yaml with your registry
sed -i.bak "s|\${DOCKER_REGISTRY}|$DOCKER_REGISTRY|g" service.yaml

# Deploy to Knative
kubectl apply -f service.yaml

# Check status
kn service list
```

## Step 4: Test Your Functions

```bash
# Get Kourier IP and service hostname
KOURIER_IP=$(kubectl get svc kourier -n kourier-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
SERVICE_HOST=$(kubectl get ksvc python-etl-example -o jsonpath='{.status.url}' | sed 's|http://||')

# Test Python ETL function
curl -X POST http://$KOURIER_IP \
  -H "Host: $SERVICE_HOST" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {"id": "1", "name": "product a", "value": "10.5"},
      {"id": "2", "name": "product b", "value": "20.3"}
    ]
  }'
```

## Development Workflow

```bash
# 1. Make code changes
vim handler.py

# 2. Build and push
docker build -t $DOCKER_REGISTRY/python-etl-example:v1.0.1 .
docker push $DOCKER_REGISTRY/python-etl-example:v1.0.1

# 3. Update service
kn service update python-etl-example \
  --image $DOCKER_REGISTRY/python-etl-example:v1.0.1

# 4. Test immediately
curl -X POST http://$KOURIER_IP -H "Host: $SERVICE_HOST" -d '{"test": "data"}'
```

## Resources

- **Knative Documentation**: https://knative.dev/docs/
- **Core Component**: `core-components/knative-serving/README.md`
- **Function Examples**: See `python-etl-example/` and `go-api-poller/`
