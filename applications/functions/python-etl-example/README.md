# Python ETL Example - Knative Function

A Flask-based ETL (Extract, Transform, Load) function running on Knative Serving.

## What it does

This function demonstrates a typical ETL pipeline:

1. **Extract**: Receives JSON data via HTTP POST
2. **Transform**: Cleans and processes the data (normalizes, validates, aggregates)
3. **Load**: Returns transformed data (in production, would write to DB/S3)

## Files

- `handler.py` - Flask application with ETL logic
- `requirements.txt` - Python dependencies
- `Dockerfile` - Container image definition
- `service.yaml` - Knative Service manifest

## Local Testing

Test the function locally without deploying:

```bash
# Install dependencies
pip install -r requirements.txt

# Run the server
python handler.py

# In another terminal, test it
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "items": [
      {"id": "1", "name": "test item", "value": "10.5"},
      {"id": "2", "name": "another item", "value": "20.3"}
    ]
  }'
```

## Deployment

### Build and Push

```bash
# Set your Docker registry
export DOCKER_REGISTRY=docker.io/yourusername

# Build the image
docker build -t $DOCKER_REGISTRY/python-etl-example:v1.0.0 .

# Push to registry
docker push $DOCKER_REGISTRY/python-etl-example:v1.0.0
```

### Deploy to Knative

```bash
# Update service.yaml with your registry
sed -i "s|\${DOCKER_REGISTRY}|$DOCKER_REGISTRY|g" service.yaml

# Deploy
kubectl apply -f service.yaml

# Check status
kn service list
kn service describe python-etl-example
```

## Testing the Deployed Function

```bash
# Get Kourier IP and service hostname
KOURIER_IP=$(kubectl get svc kourier -n kourier-system -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
SERVICE_HOST=$(kubectl get ksvc python-etl-example -o jsonpath='{.status.url}' | sed 's|http://||')

# Test with sample data
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

## Expected Output

```json
{
  "status": "success",
  "timestamp": "2025-12-31T12:00:00.000000",
  "records_processed": 2,
  "data": {
    "items": [
      {
        "id": "1",
        "name": "PRODUCT A",
        "value": 10.5,
        "processed": true
      },
      {
        "id": "2",
        "name": "PRODUCT B",
        "value": 20.3,
        "processed": true
      }
    ],
    "metadata": {
      "transformed_at": "2025-12-31T12:00:00.000000",
      "version": "1.0"
    }
  }
}
```

## Auto-Scaling

The service is configured to:
- Scale to zero when idle (saves resources)
- Scale up to 5 replicas under load
- Target 100 concurrent requests per pod

Watch it scale:

```bash
# Watch pods
kubectl get pods -w

# In another terminal, generate load
while true; do
  curl -X POST http://$KOURIER_IP \
    -H "Host: $SERVICE_HOST" \
    -H "Content-Type: application/json" \
    -d '{"test": "data"}'
  sleep 0.1
done
```

## Scheduled Execution

To run this ETL on a schedule, create a CronJob:

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: etl-daily
spec:
  schedule: "0 2 * * *"  # Run at 2 AM daily
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: trigger
            image: curlimages/curl:latest
            env:
            - name: KOURIER_IP
              value: "YOUR_KOURIER_IP"
            - name: SERVICE_HOST
              value: "python-etl-example.default.example.com"
            command:
            - /bin/sh
            - -c
            - |
              curl -X POST http://$KOURIER_IP \
                -H "Host: $SERVICE_HOST" \
                -H "Content-Type: application/json" \
                -d '{"source": "scheduled-job"}'
          restartPolicy: OnFailure
```

## Production Enhancements

For production use, consider:

1. **Database Integration**: Add PostgreSQL/MongoDB client
   ```python
   import psycopg2
   # Connect and write transformed data
   ```

2. **S3 Storage**: Use boto3 for AWS S3 or MinIO
   ```python
   import boto3
   s3 = boto3.client('s3')
   s3.put_object(Bucket='data', Key='output.json', Body=json.dumps(result))
   ```

3. **Secrets Management**: Use Kubernetes secrets
   ```yaml
   env:
   - name: DB_PASSWORD
     valueFrom:
       secretKeyRef:
         name: db-credentials
         key: password
   ```

4. **Error Handling**: Add retry logic and dead letter queues

5. **Monitoring**: Add structured logging and custom metrics

## Customization

Edit `handler.py` to implement your specific ETL logic:

```python
def transform_data(data):
    # Your custom transformation logic here
    # Examples:
    # - Call external APIs
    # - Validate against schemas
    # - Aggregate data
    # - Filter records
    # - Join with other data sources
    pass
```

## Troubleshooting

### Function won't start

```bash
# Check pod logs
kubectl logs -l serving.knative.dev/service=python-etl-example

# Check events
kubectl get events --sort-by='.lastTimestamp' | grep python-etl
```

### Image pull errors

```bash
# Make sure you're logged into your registry
docker login $DOCKER_REGISTRY

# Verify image was pushed
docker images | grep python-etl
```

## Resources

- **Flask Documentation**: https://flask.palletsprojects.com/
- **Knative Docs**: https://knative.dev/docs/
- **Python Best Practices**: https://docs.python-guide.org/
