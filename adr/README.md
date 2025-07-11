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
