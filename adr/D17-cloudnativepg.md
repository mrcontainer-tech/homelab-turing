# D17: CloudNativePG Operator for PostgreSQL

**Date**: 2026-04-03
**Status**: Proposed
**Decision**: Adopt CloudNativePG as the PostgreSQL operator
**Context**: Need a managed PostgreSQL solution for Harbor and future applications

---

## Problem Statement

The cluster currently has no managed PostgreSQL solution:
- Harbor uses an internal PostgreSQL container with a hardcoded password (`changeit`)
- VISION.md references an "Enterprise Postgres Operator" that was never deployed
- MariaDB is deployed but PostgreSQL is needed for Harbor and potentially Grafana

A Kubernetes-native PostgreSQL operator would provide automated failover, backup management, and declarative database lifecycle management.

## Goals

- Managed PostgreSQL with automated failover and recovery
- Declarative, GitOps-compatible configuration
- ARM64 support for Raspberry Pi CM4 nodes
- Integration with Longhorn storage
- Minimal resource footprint suitable for homelab

---

## Options Considered

### Option 1: CloudNativePG

CNCF Sandbox project (accepted January 2025) — community-owned, vendor-neutral PostgreSQL operator.

**Pros**:
- CNCF project — aligns with the cluster's CNCF ecosystem philosophy
- Native ARM64 support (multi-arch images)
- Declarative CRDs for cluster lifecycle management
- Built-in backup to S3 (can reuse existing Longhorn S3 backup infrastructure)
- Active development (7,700+ GitHub stars, 132M+ downloads)
- Vendor-neutral governance
- Longhorn integration documented — set replicas to 1, rely on PostgreSQL replication

**Cons**:
- CNCF Sandbox (not yet graduated)
- Relatively newer than alternatives (April 2022)
- Resource requirements may be tight on CM4 (minimum ~32Mi RAM per instance)

---

### Option 2: Zalando Postgres Operator

Mature Patroni-based operator developed by Zalando.

**Pros**:
- Battle-tested at Zalando scale
- Patroni-based HA with proven track record
- Connection pooling via PgBouncer integration

**Cons**:
- Declining maintenance activity recently
- Controlled by single company (Zalando)
- More complex architecture (Patroni dependency)
- ARM64 support less documented

---

### Option 3: CrunchyData PGO (Postgres Operator)

Feature-rich operator by Crunchy Data with PgBackRest and PgBouncer integration.

**Pros**:
- Most mature option (since March 2017)
- Feature-rich: backup, monitoring, connection pooling built-in
- Active maintenance with responsive team
- Good documentation

**Cons**:
- Controlled by single company (Crunchy Data)
- Heavier resource footprint
- More complex than CloudNativePG
- ARM64 support varies by component

---

### Option 4: Keep Current Setup (No Operator)

Continue with Harbor's internal PostgreSQL and MariaDB.

**Pros**:
- No additional components to manage
- Already working

**Cons**:
- No automated failover or recovery
- Hardcoded passwords in values files
- No backup management
- Not declarative or GitOps-friendly

---

## Comparison Matrix

| Criteria               | CloudNativePG    | Zalando          | CrunchyData PGO  | No Operator |
|------------------------|------------------|------------------|-------------------|-------------|
| CNCF Status            | Sandbox          | None             | None              | N/A         |
| ARM64 Support          | Native           | Partial          | Partial           | N/A         |
| Resource Footprint     | Low-Medium       | Medium           | Medium-High       | Lowest      |
| Maintenance Activity   | Very Active      | Declining        | Active            | N/A         |
| Vendor Neutrality      | Community-owned  | Zalando          | Crunchy Data      | N/A         |
| Longhorn Integration   | Documented       | Possible         | Possible          | N/A         |
| GitOps Friendly        | Excellent        | Good             | Good              | Poor        |
| Complexity             | Low              | Medium           | High              | None        |

---

## Decision

**Option 1: CloudNativePG** is recommended for the following reasons:

1. **CNCF alignment**: Fits the project's philosophy of showcasing CNCF ecosystem tools
2. **ARM64 native**: Multi-arch images work on Raspberry Pi CM4 nodes
3. **Lightweight**: Minimal resource overhead suitable for homelab constraints
4. **Vendor-neutral**: Community-owned, not controlled by a single company
5. **Longhorn compatible**: Documented integration pattern (single Longhorn replica + PostgreSQL replication)
6. **Active development**: Strong community momentum and clear roadmap

---

## Implementation Plan

1. Vendor the CloudNativePG Helm chart into `core-components/cloudnativepg/chart/`
2. Configure operator with ARM64-appropriate resource limits
3. Create a PostgreSQL `Cluster` resource for Harbor:
   - Replace Harbor's internal database with external CloudNativePG-managed instance
   - Update `applications/harbor/values.yaml` to use external database
   - Store credentials via external-secrets (links to Priority 1 improvements)
4. Configure automated backups to S3 (reuse existing Longhorn backup bucket)
5. Document setup in `core-components/cloudnativepg/README.md`

### Resource Estimation (CM4)

```yaml
resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
  limits:
    memory: "256Mi"
    cpu: "500m"
```

---

## References

- [CloudNativePG Documentation](https://cloudnative-pg.io/)
- [CloudNativePG CNCF Sandbox](https://www.cncf.io/projects/cloudnativepg/)
- [CloudNativePG + Longhorn Integration](https://medium.com/@camphul/cloudnative-pg-in-the-homelab-with-longhorn-b08c40b85384)
- [Comparing Kubernetes Postgres Operators](https://blog.palark.com/cloudnativepg-and-other-kubernetes-operators-for-postgresql/)
