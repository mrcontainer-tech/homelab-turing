# Knative Serving - Serverless Platform

Knative Serving is a CNCF graduated project that provides a Kubernetes-based platform for deploying and running serverless workloads.

## Architecture

- **Knative Serving**: Core serverless runtime
- **Kourier**: Lightweight ingress/networking layer
- **Custom Resource Definitions**: Service, Route, Revision, Configuration

## Features

- **Scale-to-zero**: Automatically scales down to zero when idle
- **Auto-scaling**: Request-based autoscaling (KPA) and HPA support
- **Traffic splitting**: Blue/green deployments, canary releases
- **Revision management**: Automatic versioning of deployments
- **CloudEvents**: Standard event format for event-driven architectures

## Quick Start

### 1. Install Knative CLI

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

### 2. Deploy a Service

```bash
# Deploy a container
kn service create hello \
  --image gcr.io/knative-samples/helloworld-go \
  --port 8080 \
  --env TARGET=World

# List services
kn service list

# Get service URL
kn service describe hello -o url
```

### 3. Test the Service

```bash
# Get the service URL
URL=$(kn service describe hello -o url)

# Test it
curl $URL
```

## Accessing Services

Knative services are exposed through Kourier, which creates LoadBalancer services. In your homelab:

```bash
# Check Kourier service
kubectl get svc kourier -n kourier-system

# Get the external IP (provided by MetalLB)
KOURIER_IP=$(kubectl get svc kourier -n kourier-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

# Access a service
curl -H "Host: hello.default.example.com" http://$KOURIER_IP
```

## Development Workflow

### Option 1: Using kn CLI

```bash
# Create service
kn service create my-function \
  --image docker.io/username/my-function:latest \
  --port 8080

# Update service
kn service update my-function \
  --image docker.io/username/my-function:v2

# Watch rollout
kn service describe my-function
```

### Option 2: Using YAML Manifests

```yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: my-function
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/min-scale: "0"
        autoscaling.knative.dev/max-scale: "10"
    spec:
      containers:
      - image: docker.io/username/my-function:latest
        ports:
        - containerPort: 8080
        env:
        - name: TARGET
          value: "World"
```

Apply with:
```bash
kubectl apply -f service.yaml
```

## Auto-scaling Configuration

Control scaling behavior with annotations:

```yaml
annotations:
  # Minimum number of replicas (0 = scale to zero)
  autoscaling.knative.dev/min-scale: "0"
  
  # Maximum number of replicas
  autoscaling.knative.dev/max-scale: "10"
  
  # Target concurrency per pod
  autoscaling.knative.dev/target: "100"
  
  # Scaling class (kpa or hpa)
  autoscaling.knative.dev/class: "kpa"
  
  # Scale down delay
  autoscaling.knative.dev/scale-down-delay: "0s"
  
  # Window for stable mode
  autoscaling.knative.dev/window: "60s"
```

## Traffic Splitting

Knative supports advanced traffic management:

```bash
# Deploy new revision
kn service update my-function --image new-image:v2 --revision-name v2

# Split traffic: 90% to v1, 10% to v2 (canary)
kn service update my-function \
  --traffic v1=90,v2=10

# Blue/green: send all traffic to v2
kn service update my-function \
  --traffic v2=100

# Rollback
kn service update my-function \
  --traffic v1=100
```

## Monitoring

### View Logs

```bash
# All replicas
kn service logs my-function

# Follow logs
kn service logs my-function -f

# Specific revision
kubectl logs -n default -l serving.knative.dev/revision=my-function-v2
```

### Check Status

```bash
# Service status
kn service describe my-function

# Revisions
kn revision list

# Routes
kn route list
```

## Knative vs Traditional Kubernetes

| Feature | Traditional K8s | Knative |
|---------|----------------|---------|
| Deployment | Manual Deployment + Service | Single Service resource |
| Scaling | Manual HPA setup | Built-in auto-scaling |
| Scale-to-zero | Not supported | Automatic |
| Revisions | Manual versioning | Automatic |
| Traffic splitting | Complex (Ingress rules) | Built-in |
| Cold start | N/A | ~1-2 seconds |

## Integration with Your Homelab

Knative integrates with your existing infrastructure:

- ✅ **MetalLB**: Kourier LoadBalancer uses MetalLB IPs
- ✅ **Traefik**: Can expose Knative services via Traefik IngressRoute
- ✅ **cert-manager**: Use for TLS certificates
- ✅ **external-dns**: Auto-create DNS records for services
- ✅ **Prometheus**: Knative exports metrics
- ✅ **ArgoCD**: Deploy services via GitOps

## Best Practices

1. **Set resource limits**: Prevent runaway containers
   ```yaml
   resources:
     limits:
       memory: "512Mi"
       cpu: "1000m"
     requests:
       memory: "256Mi"
       cpu: "500m"
   ```

2. **Use health checks**: Improve reliability
   ```yaml
   livenessProbe:
     httpGet:
       path: /health
       port: 8080
   readinessProbe:
     httpGet:
       path: /ready
       port: 8080
   ```

3. **Configure appropriate scaling**: Match your workload
   ```yaml
   annotations:
     autoscaling.knative.dev/min-scale: "1"  # No cold starts
     autoscaling.knative.dev/max-scale: "5"  # Limit resource usage
   ```

4. **Use proper versioning**: Tag images, don't use `latest`
   ```yaml
   image: docker.io/username/app:v1.0.0  # ✅ Good
   # image: docker.io/username/app:latest  # ❌ Avoid
   ```

## Troubleshooting

### Service won't become ready

```bash
# Check service status
kn service describe my-function

# Check pod logs
kubectl logs -n default -l serving.knative.dev/service=my-function

# Check events
kubectl get events -n default --sort-by='.lastTimestamp'
```

### Can't access service

```bash
# Check Kourier is running
kubectl get pods -n kourier-system

# Check service has external address
kn service list

# Verify MetalLB assigned IP to Kourier
kubectl get svc kourier -n kourier-system
```

### Service scaling issues

```bash
# Check autoscaler logs
kubectl logs -n knative-serving -l app=autoscaler

# Check metrics
kubectl get podautoscalers -n default
```

## Resources

- **Knative Docs**: https://knative.dev/docs/
- **Knative GitHub**: https://github.com/knative/serving
- **CNCF Project**: https://www.cncf.io/projects/knative/
- **Examples**: https://knative.dev/docs/samples/

## Configuration Files

The following files configure Knative Serving:

- `serving-crds.yaml`: Custom Resource Definitions
- `serving-core.yaml`: Core Knative components
- `kourier.yaml`: Networking layer
- `config.yaml`: Domain and network configuration
- `namespace.yaml`: Required namespaces

## Security Notes

Knative services run with default Kubernetes RBAC. For production:

1. Use private container registries
2. Store secrets in Kubernetes Secrets
3. Configure network policies
4. Enable Pod Security Standards
5. Use service mesh (Istio) for mTLS
