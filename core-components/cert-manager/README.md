# Cert Manager

To speed things along, I created an IAM user with least privilege permissions to create Route53 records. The secret itself I created using the AWS CLI and Kubectl. Further improvement to prevent this is to use something like Identity webhook or IAM Roles Anywhere.

## Prerequisites

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
kubectl create secret generic certmanager --namespace cert-manager --from-file credentials-certmanager
```