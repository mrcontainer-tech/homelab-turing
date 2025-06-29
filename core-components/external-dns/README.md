# Manual steps for External DNS

## Prerequisites

- A Kubernetes cluster with the ExternalDNS controller installed.
- An AWS account with the necessary permissions to create and manage Route53 records.

## Steps

1. Create an IAM user with the necessary permissions to create and manage Route53 records.

2. Create an AWS credentials file with the access key ID and secret access key for the IAM user.

```
[default]
aws_access_key_id = YOUR_ACCESS_KEY_ID
aws_secret_access_key = YOUR_SECRET_ACCESS_KEY
```

3. Create a Kubernetes secret with the AWS credentials file.

```
kubectl create secret generic external-dns 
  --namespace external-dns --from-file credentials
```

