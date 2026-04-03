# D16: Harbor Migration to ARM64

**Date**: 2026-04-03
**Status**: Accepted
**Decision**: Option 1 - Migrate Harbor to ARM64 using Bitnami Helm chart
**Context**: Node5 (AMD64) has been decommissioned; Harbor was exclusively scheduled on node5

---

## Problem Statement

Harbor was deployed exclusively on node5 (Dell XPS, x86_64) using the official goharbor Helm chart with `nodeSelector: kubernetes.io/hostname: node5` on all components. With node5 decommissioned, Harbor cannot run in the cluster.

The remaining cluster nodes are all Raspberry Pi CM4s (ARM64) on the Turing Pi v2.5. Harbor needs to either migrate to ARM64 or be removed.

## Goals

- Restore Harbor container registry functionality in the cluster
- Maintain proxy cache capabilities (DockerHub, GHCR, Quay)
- Keep Kyverno policy enforcement (D10) operational
- Stay within Raspberry Pi CM4 resource constraints

---

## Options Considered

### Option 1: Bitnami Harbor Helm Chart (ARM64 native)

Switch from the official goharbor chart to the Bitnami Harbor chart, which provides multi-arch images including ARM64.

**Pros**:
- Native ARM64 support (multi-arch images)
- Already evaluated in D2 — original decision was to use Bitnami for ARM compatibility
- Consistent with Bitnami chart patterns used elsewhere
- Active maintenance and security updates

**Cons**:
- Different chart structure than goharbor — migration effort for values.yaml
- Bitnami charts can be opinionated about defaults
- Raspberry Pi CM4 resource constraints may limit performance

---

### Option 2: Official goharbor Chart with Multi-Arch Images

Keep the current goharbor chart but remove nodeSelectors, relying on Harbor's multi-arch image support.

**Pros**:
- Minimal configuration change (just remove nodeSelectors)
- No chart migration needed

**Cons**:
- Official goharbor images have inconsistent ARM64 support
- Some components (Trivy, Notary) may lack ARM64 images
- Higher risk of runtime failures on ARM64

---

### Option 3: Remove Harbor Until New Hardware

Remove Harbor from the cluster temporarily until new x86_64/ARM64 hardware (Jetson Orin NX, Turing RK1) is added.

**Pros**:
- No ARM compatibility risk
- Frees up cluster resources

**Cons**:
- Breaks Kyverno policy enforcement (D10) — all image pulls would fail policy checks
- Loses proxy cache benefits (bandwidth, rate limit avoidance)
- Unknown timeline for new hardware

---

## Comparison Matrix

| Criteria                | Bitnami Chart | goharbor Chart | Remove Harbor |
|-------------------------|---------------|----------------|---------------|
| ARM64 Support           | Native        | Partial        | N/A           |
| Migration Effort        | Medium        | Low            | Low           |
| Risk                    | Low           | Medium-High    | High (policy) |
| Resource Usage on Pi    | Medium        | Medium         | None          |
| Kyverno Policy (D10)    | Maintained    | Maintained     | Broken        |

---

## Decision

**Option 1: Bitnami Harbor Helm chart** is the recommended path forward for the following reasons:

1. **Proven ARM64 support**: Bitnami provides tested multi-arch images, already validated in D2
2. **Kyverno continuity**: Maintains the Harbor registry enforcement policy (D10)
3. **Proxy cache preservation**: Cluster continues to benefit from cached images and rate limit avoidance
4. **Alignment with D2**: This was the original intent — D2 specifically chose Bitnami for ARM compatibility

As an interim step, the node5 nodeSelectors have been removed from the current goharbor chart values to unblock scheduling on ARM64 nodes. The full migration to the Bitnami chart should follow.

---

## Implementation Plan

1. Remove all node5 nodeSelectors from `applications/harbor/values.yaml` (done)
2. Vendor the Bitnami Harbor Helm chart into `applications/harbor/chart/`
3. Migrate values.yaml to Bitnami chart format
4. Test with `helm template` to validate ARM64 compatibility
5. Verify proxy cache configurations carry over
6. Confirm Kyverno policies (D10) work with the new deployment

---

## References

- D2: Adopt Bitnami Harbor Helm chart for ARM support
- D9: Use Harbor as container registry / proxy cache
- D10: Use Kyverno to enforce Harbor repository usage
- [Bitnami Harbor Chart](https://github.com/bitnami/charts/tree/main/bitnami/harbor)
- [Harbor Official Chart](https://github.com/goharbor/harbor-helm)
