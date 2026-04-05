# DECISION LOG

A concise, lightweight decision log inspired by Architectural Decision Records (ADRs). Created it showcase a (simple) way of documenting decisions made in a semi-complex system.

## Statuses

| Status       | Meaning                                                  |
| ------------ | -------------------------------------------------------- |
| Accepted     | Decision is approved and being implemented or completed  |
| Proposed     | Decision is drafted but not yet approved                 |
| Superseded   | Decision has been replaced by a newer decision           |
| Deprecated   | Decision is no longer relevant                           |

## Table of Decisions

| ID  | Date       | Status     | Decision                                                                  | Notes                                                            |
| --- | ---------- | ---------- | ------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| D1  | 2025-07-11 | Accepted   | Begin using this decision log to track key decisions                      | Created initial template for recording decisions                 |
| D2  | 2025-07-11 | Superseded | Adopt Bitnami Harbor Helm chart for ARM support                           | Superseded by D20                                                |
| D3  | 2025-07-11 | Accepted   | Split into 2 appsets: core-components and applications                    | Nice split between core-components and apps                      |
| D4  | 2025-07-11 | Accepted   | Use AWS for Route53 (cert-manager & external-dns)                         | Have a way to get proper tls and fqdns                           |
| D5  | 2025-07-11 | Accepted   | Use ArgoCD for GitOps                                                     | Looked into flux but went for ArgoCD because personal preference |
| D6  | 2025-07-11 | Superseded | Use Raspberry Pi OS                                                       | Superseded by D16 (Talos migration)                              |
| D7  | 2025-07-11 | Superseded | Use k3s as Kubernetes distro                                              | Superseded by D16 (Talos migration)                              |
| D8  | 2025-07-11 | Accepted   | Use homepage as a landingpage for my Homelab                              | Nice way of having a customizable homepage                       |
| D9  | 2025-07-11 | Accepted   | Use Harbor as a way to pull in images that are used in the cluster        | Excluded are the images used by K3s                              |
| D10 | 2025-07-11 | Accepted   | Use Kyverno as a way to force to use the Harbor repository                | Really great way of doing policy-as-code                         |
| D11 | 2025-12-31 | Deprecated | Implement Knative                                                         | Testing out Knative and actually get some code up and running    |
| D12 | 2025-12-31 | Accepted   | Fix coreDNS issue                                                         | Issues with Tailscale DNS                                        |
| D13 | 2025-12-31 | Accepted   | Improve home media setup                                                  | Working JellyFin, Jellyseerr, Sonarr, etc.                      |
| D14 | 2026-01-01 | Superseded | Media app credentials managed manually, not committed to Git              | Superseded by external-secrets integration with AWS Secrets Manager |
| D15 | 2026-01-20 | Accepted   | Implement encrypted DNS using dnscrypt-proxy                              | Multi-provider DoH/DNSCrypt support. See [D15-encrypted-dns.md](D15-encrypted-dns.md) |
| D16 | 2026-04-03 | Proposed   | Migrate from k3s to Talos Linux                                           | Immutable OS, reduced maintenance, new hardware support. Supersedes D6, D7. See [D16-talos-migration.md](D16-talos-migration.md) |
| D17 | 2026-04-03 | Proposed   | Adopt CloudNativePG for managed PostgreSQL                                | CNCF Sandbox, ARM64 native, replace Harbor internal DB. See [D17-cloudnativepg.md](D17-cloudnativepg.md) |
| D18 | 2026-04-03 | Accepted   | Use Tekton as CI system                                                   | Kubernetes-native, CNCF Incubating. See [D18-ci-system.md](D18-ci-system.md) |
| D19 | 2026-04-03 | Accepted   | OIDC federation for AWS authentication                                    | Replace static IAM keys with IRSA-style keyless auth. See [D19-oidc-aws-authentication.md](D19-oidc-aws-authentication.md) |
| D20 | 2026-04-03 | Accepted   | Migrate Harbor to ARM64                                                   | Node5 decommissioned; switch to Bitnami chart for ARM support. See [D20-harbor-on-arm.md](D20-harbor-on-arm.md) |
| D21 | 2026-04-03 | Proposed   | Implement layered cluster security                                        | VAP + Kyverno + Tetragon + PSS + NetworkPolicies. See [D21-cluster-security.md](D21-cluster-security.md) |
| D22 | 2026-04-04 | Accepted   | Tailscale remote access for homelab services                              | Hybrid: subnet router + Tailscale Ingress for priority services. See [D22-tailscale-remote-access.md](D22-tailscale-remote-access.md) |
