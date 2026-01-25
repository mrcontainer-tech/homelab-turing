# Kyverno

Kubernetes Native Policy Management - policy engine for validating, mutating, and generating Kubernetes resources.

## Chart Info

- **Chart**: kyverno/kyverno v3.6.2
- **App**: Kyverno v1.16.2
- **Source**: https://kyverno.io/

## Components

Kyverno deploys several controllers:

| Controller | Purpose |
|------------|---------|
| Admission Controller | Validates and mutates resources during admission |
| Background Controller | Processes policies on existing resources |
| Cleanup Controller | Handles cleanup policies |
| Reports Controller | Generates policy reports |

## Usage

Create policies using Kyverno CRDs:

```yaml
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: require-labels
spec:
  validationFailureAction: Enforce
  rules:
    - name: check-for-labels
      match:
        resources:
          kinds:
            - Pod
      validate:
        message: "Label 'app' is required"
        pattern:
          metadata:
            labels:
              app: "?*"
```

## Related

- [Policy Reporter](../policy-reporter/) - UI for viewing policy reports
