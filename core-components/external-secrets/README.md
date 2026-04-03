# External Secrets Operator

Syncs secrets from external secret management systems (e.g., AWS Secrets Manager) into Kubernetes Secrets.

## Overview

The External Secrets Operator (ESO) is deployed via vendored Helm chart (v0.18.1). It watches for `ExternalSecret` resources and creates corresponding Kubernetes Secrets by fetching data from a configured backend.

## Architecture

```
AWS Secrets Manager → ClusterSecretStore → ExternalSecret → Kubernetes Secret
```

## Setup

### Prerequisites

1. **AWS IAM User** with `secretsmanager:GetSecretValue` and `secretsmanager:ListSecrets` permissions
2. **Kubernetes Secret** containing the AWS credentials:

```bash
kubectl create secret generic aws-secret-manager-credentials \
  --namespace external-secrets \
  --from-literal=access-key-id=<AWS_ACCESS_KEY_ID> \
  --from-literal=secret-access-key=<AWS_SECRET_ACCESS_KEY>
```

### ClusterSecretStore

A `ClusterSecretStore` named `aws-secrets-manager` is defined in `manifests/cluster-secret-store.yaml`. This provides cluster-wide access to AWS Secrets Manager.

### Creating ExternalSecrets

To sync a secret from AWS Secrets Manager into a Kubernetes namespace:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: my-app-secret
  namespace: my-namespace
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: ClusterSecretStore
  target:
    name: my-app-secret
  data:
    - secretKey: password
      remoteRef:
        key: my-app/password
```

## Components

| Component | Purpose |
|-----------|---------|
| Operator | Watches ExternalSecret CRs and reconciles Kubernetes Secrets |
| Webhook | Validates ExternalSecret and SecretStore resources |
| Cert Controller | Manages TLS certificates for the webhook |

## Related Decisions

- D14: Media app credentials managed manually until external-secrets backend is configured
