# DECISION LOG

A concise, lightweight decision log inspired by Architectural Decision Records (ADRs). Created it showcase a (simple) way of documenting decisions made in a semi-complex system.

## Table of Decisions

| ID  | Date       | Decision                                                                  | Notes                                                            |
| --- | ---------- | ------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| D1  | 2025-07-11 | Begin using this decision log to track key decisions                      | Created initial template for recording decisions                 |
| D2  | 2025-07-11 | Adopt Bitnami Harbor Helm chart for ARM support                           | Ensures multi-arch compatibility on Raspberry Pi                 |
| D3  | 2025-07-11 | Retro: Split into 2 appsets: core-components and applications             | Nice split between core-components and apps                      |
| D4  | 2025-07-11 | Retro: Use AWS for Route53 (cert-manager & external-dns)                  | Have a way to get proper tls and fqdns                           |
| D5  | 2025-07-11 | Retro: Use ArgoCD for GitOps                                              | Looked into flux but went for ArgoCD because personal preference |
| D6  | 2025-07-11 | Retro: Use Raspbarry Pi OS over for example Talos                         | Still want to have the ability to reach the node os              |
| D7  | 2025-07-11 | Retro: Use k3s as Kubernetes distro                                       | Went for K3s as personal preference/experience                   |
| D8  | 2025-07-11 | Retro: Use homepage as a landingpage for my Homelab                       | Nice way of having a customizable homepage                       |
| D9  | 2025-07-11 | Retro: Use Harbor as a way to pull in images that are used in the cluster | Excluded are the images used by K3s                              |
| D10 | 2025-07-11 | Retro: Use Kyverno as a way to force to use the Harbor repository         | Really great way of doing policy-as-code                         |
| D11 | 2025-12-31 | Implement Knative                                                         | Testing out Knative and actually get some code up and running.   |
| D12 | 2025-12-31 | Fix coreDNS issue                                                         | Issues with Tailscale DNS                                        |
| D13 | 2025-12-31 | Improve home media setup                                                  | Working JellyFin, Jellyseerr, Sonarr, etc.                       |
| D14 | 2026-01-01 | Media app credentials managed manually, not committed to Git             | Created `-creds.yaml` files (gitignored) with init containers for API key persistence. Applied manually via kubectl until external-secrets backend is configured. |
| D15 | 2026-01-20 | Implement encrypted DNS using dnscrypt-proxy                             | Multi-provider DoH/DNSCrypt support. See [D15-encrypted-dns.md](D15-encrypted-dns.md) |
| D16 | 2026-04-03 | Migrate from k3s to Talos Linux                                         | Immutable OS, reduced maintenance, new hardware support. Supersedes D6, D7. See [D16-talos-migration.md](D16-talos-migration.md) |
| D17 | 2026-04-03 | Adopt CloudNativePG for managed PostgreSQL                               | CNCF Sandbox, ARM64 native, replace Harbor internal DB. See [D17-cloudnativepg.md](D17-cloudnativepg.md) |
| D18 | 2026-04-03 | Use Tekton as CI system                                                  | Kubernetes-native, CNCF Incubating. See [D18-ci-system.md](D18-ci-system.md) |
| D19 | 2026-04-03 | OIDC federation for AWS authentication                                   | Replace static IAM keys with IRSA-style keyless auth. See [D19-oidc-aws-authentication.md](D19-oidc-aws-authentication.md) |
| D20 | 2026-04-03 | Migrate Harbor to ARM64                                                  | Node5 decommissioned; switch to Bitnami chart for ARM support. See [D20-harbor-on-arm.md](D20-harbor-on-arm.md) |
| D21 | 2026-04-03 | Implement layered cluster security                                       | VAP + Kyverno + Tetragon + PSS + NetworkPolicies. See [D21-cluster-security.md](D21-cluster-security.md) |
