# Go API Poller - Knative Function

A Go-based HTTP server that polls external APIs and processes responses, running on Knative Serving.

## What it does

This function:

1. **Polls**: Makes HTTP requests to external APIs
2. **Processes**: Parses and validates JSON responses
3. **Returns**: Structured results with status and timing

## Files

- `handler.go` - Go HTTP server with polling logic
- `go.mod` - Go module definition
- `Dockerfile` - Multi-stage container build
- `service.yaml` - Knative Service manifest

## Local Testing

Test the function locally:

```bash
# Run the server
go run handler.go

# In another terminal, test health check
curl http://localhost:8080/

# Test with API polling
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://api.github.com/users/octocat",
    "method": "GET"
  }'
```

## Deployment

### Build and Push

```bash
# Set your Docker registry
export DOCKER_REGISTRY=docker.io/yourusername

# Build the image
docker build -t $DOCKER_REGISTRY/go-api-poller:v1.0.0 .

# Push to registry
docker push $DOCKER_REGISTRY/go-api-poller:v1.0.0
```

### Deploy to Knative

```bash
# Update service.yaml with your registry
sed -i "s|\${DOCKER_REGISTRY}|$DOCKER_REGISTRY|g" service.yaml

# Deploy
kubectl apply -f service.yaml

# Check status
kn service list
kn service describe go-api-poller
```

## Testing the Deployed Function

```bash
# Get Kourier IP and service hostname
KOURIER_IP=$(kubectl get svc kourier -n kourier-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
SERVICE_HOST=$(kubectl get ksvc go-api-poller -o jsonpath='{.status.url}' | sed 's|http://||')

# Test with default endpoint (GitHub Zen)
curl -X POST http://$KOURIER_IP \
  -H "Host: $SERVICE_HOST"

# Test with custom endpoint
curl -X POST http://$KOURIER_IP \
  -H "Host: $SERVICE_HOST" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://api.github.com/users/octocat",
    "method": "GET"
  }'

# Test with headers
curl -X POST http://$KOURIER_IP \
  -H "Host: $SERVICE_HOST" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://api.example.com/data",
    "method": "GET",
    "headers": {
      "Authorization": "Bearer token123",
      "Accept": "application/json"
    }
  }'
```

## Expected Output

```json
{
  "success": true,
  "status_code": 200,
  "response_body": {
    "login": "octocat",
    "id": 583231,
    "name": "The Octocat"
  },
  "poll_time": "2025-12-31T12:00:00Z",
  "duration": "250ms"
}
```

## Use Cases

- **Health Monitoring**: Poll service health endpoints
- **API Status Checks**: Monitor third-party API availability
- **Data Collection**: Fetch data from external sources
- **Webhook Triggers**: Check for changes and trigger actions
- **Integration Testing**: Verify external service behavior

## Scheduled Polling

Run this function on a schedule to continuously monitor APIs:

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: api-health-check
spec:
  schedule: "*/5 * * * *"  # Every 5 minutes
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: poller
            image: curlimages/curl:latest
            env:
            - name: KOURIER_IP
              value: "YOUR_KOURIER_IP"
            - name: SERVICE_HOST
              value: "go-api-poller.default.example.com"
            command:
            - /bin/sh
            - -c
            - |
              curl -X POST http://$KOURIER_IP \
                -H "Host: $SERVICE_HOST" \
                -H "Content-Type: application/json" \
                -d '{"url": "https://api.example.com/health"}'
          restartPolicy: OnFailure
```

## Production Enhancements

1. **Add Alerting**: Integrate with Slack, Discord, or PagerDuty
   ```go
   if !result.Success {
       sendAlert("API is down!")
   }
   ```

2. **Store Results**: Write to PostgreSQL, InfluxDB, or Prometheus
   ```go
   db.Exec("INSERT INTO api_checks (url, status, duration) VALUES (?, ?, ?)",
       req.URL, result.StatusCode, result.Duration)
   ```

3. **Authentication**: Add OAuth2 or JWT token management
   ```go
   token := getAccessToken()
   httpReq.Header.Set("Authorization", "Bearer "+token)
   ```

4. **Retry Logic**: Implement exponential backoff
   ```go
   for attempt := 0; attempt < maxRetries; attempt++ {
       resp, err = client.Do(httpReq)
       if err == nil { break }
       time.Sleep(time.Duration(attempt) * time.Second)
   }
   ```

5. **Metrics**: Export custom Prometheus metrics
   ```go
   apiCallCounter.Inc()
   apiDuration.Observe(duration.Seconds())
   ```

## Customization

Edit `handler.go` to add custom logic:

```go
func pollAPI(req PollRequest) PollResult {
    // ... existing code ...
    
    // Add custom actions based on response
    if result.Success {
        // Trigger webhook
        // Store in database
        // Send notification
    } else {
        // Send alert
        // Log error
        // Retry logic
    }
    
    return result
}
```

## Environment Variables

Configure via service.yaml:

```yaml
env:
- name: API_ENDPOINT
  value: "https://api.example.com/status"
- name: API_TOKEN
  valueFrom:
    secretKeyRef:
      name: api-credentials
      key: token
```

## Troubleshooting

### Function won't start

```bash
# Check pod logs
kubectl logs -l serving.knative.dev/service=go-api-poller

# Check events
kubectl get events --sort-by='.lastTimestamp' | grep go-api
```

### Build errors

```bash
# Test build locally
go build handler.go

# Check for dependency issues
go mod tidy
```

## Resources

- **Go Documentation**: https://go.dev/doc/
- **Knative Docs**: https://knative.dev/docs/
- **HTTP Client Guide**: https://pkg.go.dev/net/http
