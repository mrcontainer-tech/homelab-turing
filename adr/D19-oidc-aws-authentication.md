# D19: OIDC Federation for AWS Authentication

**Date**: 2026-04-03
**Status**: Accepted
**Decision**: Use OIDC federation (IRSA-style) for keyless AWS authentication
**Context**: Cluster workloads authenticate to AWS using static IAM access keys — replace with OIDC for improved security

---

## Problem Statement

Three components currently authenticate to AWS using static IAM access keys stored as Kubernetes Secrets:

| Component | AWS Service | Credential Secret |
|-----------|------------|-------------------|
| cert-manager | Route53 | `certmanager` (namespace: cert-manager) |
| external-dns | Route53 | `external-dns` (namespace: external-dns) |
| external-secrets | Secrets Manager | `aws-secret-manager-credentials` (namespace: external-secrets) — not yet created |

Static credentials are:
- A security risk — long-lived keys that must be manually rotated
- Operationally fragile — stored as gitignored files applied manually via kubectl
- Not GitOps-friendly — credential lifecycle is outside the Git workflow

## Goals

- Eliminate static AWS access keys from the cluster
- Authenticate workloads using short-lived, automatically rotated tokens
- Maintain compatibility with the Talos migration (D16)
- Apply consistently across all AWS-integrated components

---

## Options Considered

### Option 1: OIDC Federation (IRSA-style)

Kubernetes service account tokens are JWTs signed by the cluster. AWS IAM can be configured to trust these tokens via an OIDC Identity Provider, allowing pods to assume IAM roles without static credentials.

**Architecture**:
```
Pod → projected SA token (JWT) → AWS STS AssumeRoleWithWebIdentity → temporary credentials
```

**Setup**:
1. Host cluster OIDC discovery documents (`.well-known/openid-configuration` + JWKS) in a public S3 bucket
2. Create IAM OIDC Identity Provider pointing to the S3 bucket
3. Create IAM roles with trust policies scoped to specific ServiceAccounts
4. Deploy amazon-eks-pod-identity-webhook to inject token volumes into annotated pods

**Pros**:
- No static credentials — tokens are short-lived and auto-rotated
- Fine-grained: each ServiceAccount maps to a specific IAM role with least-privilege permissions
- Kubernetes-native — uses projected service account tokens
- Works on k3s and Talos (same OIDC setup transfers across the migration)
- Well-documented pattern used by EKS, adapted for self-managed clusters
- Pod identity webhook available as Helm chart for non-EKS clusters

**Cons**:
- Requires a publicly accessible S3 bucket for OIDC discovery documents
- Initial setup complexity (one-time)
- Pod identity webhook is an additional component to maintain
- JWKS must be updated if service account signing keys change

---

### Option 2: IAM Roles Anywhere

Uses X.509 certificates to authenticate to AWS. A certificate authority (ACM PCA or self-managed) issues certificates to workloads, which exchange them for temporary AWS credentials.

**Pros**:
- No public endpoint required
- Works in fully air-gapped environments

**Cons**:
- Higher complexity — requires certificate issuance, rotation, and a credential helper sidecar
- More moving parts (ACM PCA or custom CA + `aws-rolesanywhere-credential-helper`)
- Less community adoption for Kubernetes workloads
- Certificate rotation adds operational overhead

---

### Option 3: Keep Static Credentials

Continue using IAM access keys stored as Kubernetes Secrets.

**Pros**:
- Simple, already partially working
- No additional infrastructure needed

**Cons**:
- Long-lived credentials with no automatic rotation
- Manual lifecycle management
- Security risk if keys are leaked
- Not aligned with zero-trust / GitOps principles

---

## Comparison Matrix

| Criteria                 | OIDC Federation    | IAM Roles Anywhere | Static Keys    |
|--------------------------|--------------------|--------------------|----------------|
| Credential Lifetime      | Short (hours)      | Short (hours)      | Long (manual)  |
| Rotation                 | Automatic          | Semi-automatic     | Manual         |
| Setup Complexity         | Medium (one-time)  | High               | Low            |
| Public Endpoint Required | Yes (S3 bucket)    | No                 | No             |
| Kubernetes Native        | Yes                | No (sidecar)       | Yes            |
| Talos Compatible         | Yes                | Yes                | Yes            |
| Community Adoption       | High (EKS pattern) | Low                | High           |
| Security Posture         | Strong             | Strong             | Weak           |

---

## Decision

**Option 1: OIDC Federation** is chosen for the following reasons:

1. **Eliminates static credentials** — short-lived tokens replace long-lived access keys
2. **Kubernetes-native** — uses projected service account tokens, no sidecars needed
3. **Fine-grained IAM** — each component gets its own IAM role with least-privilege permissions
4. **Talos migration ready** — identical setup works on Talos (D16), just reuse the same signing keys and S3 bucket
5. **Proven pattern** — standard IRSA approach used across the AWS/Kubernetes ecosystem
6. **Consistent** — same authentication model for cert-manager, external-dns, and external-secrets

---

## Implementation Plan

### Phase 1: OIDC Infrastructure

1. **Extract service account signing keys** from k3s
2. **Create S3 bucket** (e.g., `homelab-turing-oidc`) in eu-west-1 with public read on OIDC paths
3. **Upload OIDC discovery documents**:
   - `.well-known/openid-configuration`
   - `oidc/v1/jwks`
4. **Configure k3s** `--service-account-issuer` to match the S3 URL
5. **Create IAM OIDC Identity Provider** in AWS pointing to the S3 bucket

### Phase 2: IAM Roles

Create three IAM roles with least-privilege trust policies:

| Role | ServiceAccount | Permissions |
|------|---------------|-------------|
| `homelab-external-secrets` | `external-secrets:external-secrets` | `secretsmanager:GetSecretValue`, `secretsmanager:ListSecrets` |
| `homelab-cert-manager` | `cert-manager:cert-manager` | Route53 `ChangeResourceRecordSets`, `ListHostedZonesByName`, `GetChange` |
| `homelab-external-dns` | `external-dns:external-dns` | Route53 `ChangeResourceRecordSets`, `ListResourceRecordSets`, `ListHostedZonesByName` |

### Phase 3: Pod Identity Webhook

1. Deploy `amazon-eks-pod-identity-webhook` via Helm into `kube-system`
2. The webhook automatically injects `AWS_ROLE_ARN`, `AWS_WEB_IDENTITY_TOKEN_FILE`, and projected token volumes into pods with annotated ServiceAccounts

### Phase 4: Migrate Components

**external-secrets** — update `ClusterSecretStore`:
```yaml
spec:
  provider:
    aws:
      service: SecretsManager
      region: eu-west-1
      auth:
        jwt:
          serviceAccountRef:
            name: external-secrets
            namespace: external-secrets
```

**cert-manager** — update `values.yaml`:
- Remove `extraEnv`, `volumes`, `volumeMounts` referencing AWS credentials file
- Add ServiceAccount annotation with IAM role ARN

**external-dns** — update `values.yaml`:
- Remove `extraVolumes`, `extraVolumeMounts` referencing AWS credentials
- Add ServiceAccount annotation with IAM role ARN

### Phase 5: Cleanup

1. Delete static credential Secrets from the cluster
2. Delete the IAM user access keys
3. Remove gitignored credential files (`credentials-certmanager`, `credentials-externaldns`)

---

## Talos Migration Continuity

When migrating to Talos (D16):
- Reuse the same service account signing keys in the Talos machine config
- S3 bucket and IAM OIDC provider remain unchanged
- ServiceAccount annotations are identical
- No changes needed to ClusterSecretStore, cert-manager, or external-dns configs

---

## References

- [AWS IRSA for Self-Hosted Kubernetes](https://reece.tech/posts/oidc-k8s-to-aws/)
- [Talos Linux IRSA Documentation](https://docs.siderolabs.com/talos/v1.7/security/iam-roles-for-service-accounts)
- [amazon-eks-pod-identity-webhook](https://github.com/aws/amazon-eks-pod-identity-webhook)
- [External Secrets — AWS Provider](https://external-secrets.io/latest/provider/aws-secrets-manager/)
- [cert-manager — Route53 with IRSA](https://cert-manager.io/docs/configuration/acme/dns01/route53/)
- [AWS IAM Roles Anywhere](https://aws.amazon.com/iam/roles-anywhere/)
- D14: Media app credentials managed manually
- D16: Talos migration
