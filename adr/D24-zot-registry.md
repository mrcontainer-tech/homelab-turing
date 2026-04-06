# D24: Replace Harbor with Zot Registry

**Date**: 2026-04-06
**Status**: Proposed
**Decision**: Replace Harbor with Zot as the cluster container registry
**Context**: Harbor is amd64-only and cannot run on the ARM64 cluster. Zot is a CNCF Sandbox project that provides image scanning, pull-through caching, and a web UI natively on ARM64.

---

## Problem Statement

Harbor has been broken since Node5 (Dell XPS, x86_64) was decommissioned. All goharbor container images (core, portal, nginx, registry, trivy, redis, database) are built exclusively for amd64. The cluster now consists entirely of Raspberry Pi CM4 nodes (ARM64).

D20 proposed migrating to the Bitnami Harbor chart for ARM64 support. Further investigation reveals that even Bitnami does not resolve the issue, as Harbor's upstream images remain amd64-only. The cluster has been without a functioning container registry for over 2 days, breaking:

- **Image scanning**: No vulnerability scanning on pulled images
- **Pull-through caching**: Direct pulls from Docker Hub, GHCR, etc. (rate limit risk)
- **Kyverno policy enforcement (D10)**: Policies that require images come through the registry are ineffective
- **Visibility**: No central dashboard showing which images the cluster uses

## Goals

- Restore a functioning container registry on the ARM64 cluster
- Provide built-in image vulnerability scanning (Trivy or equivalent)
- Re-establish pull-through caching for upstream registries (Docker Hub, GHCR, Quay, LSCR, k8s, GCR, NVCR)
- Provide a web dashboard for browsing images, viewing scan results, and understanding cluster image usage
- Maintain compatibility with Kyverno image source policies (D10)
- Fit within Raspberry Pi CM4 resource constraints (~4GB RAM per node)

---

## Options Considered

### Option 1: Zot Registry (CNCF Sandbox)

A vendor-neutral OCI-native container registry built from the ground up for security and simplicity. Single static Go binary with no external dependencies (no database, no Redis).

**Pros**:
- Native ARM64 support with specific Raspberry Pi optimizations (`commit` storage mode for RAM-constrained devices)
- Built-in Trivy scanning (embedded as a Go library, automatic background scanning)
- Pull-through caching via the Sync extension with `onDemand: true`
- Built-in web UI for browsing images, searching, viewing CVE scan results
- Single pod deployment, minimal resource footprint (~300m CPU, ~512Mi-1Gi RAM)
- Harbor Satellite (Harbor's own edge solution) embeds Zot, validating the technology
- OCI-native: supports OCI artifacts, signatures, SBOMs natively

**Cons**:
- CNCF Sandbox (less mature than Graduated projects)
- Smaller community than Harbor
- Web UI is more basic than Harbor's full dashboard
- No built-in replication to remote registries (Sync is pull-only)

---

### Option 2: Distribution / Docker Registry v2 (CNCF Sandbox)

The reference implementation for the OCI Distribution Specification. Powers Docker Hub, GHCR, and GitLab Container Registry under the hood.

**Pros**:
- Native ARM64 support (multi-arch images)
- Battle-tested core technology
- Native pull-through cache via `proxy.remoteurl`
- Very low resource footprint

**Cons**:
- No image scanning capability whatsoever
- No web UI or dashboard (API-only)
- Stagnated at CNCF Sandbox since 2021
- Cannot serve as a central scanning hub (fails primary requirement)

---

### Option 3: Dragonfly (CNCF Graduated)

A P2P-based image and file distribution system designed for large-scale infrastructure.

**Pros**:
- CNCF Graduated (highest maturity level)
- ARM64 support added in v2.4.0
- P2P distribution reduces bandwidth at scale
- Web console for cluster management

**Cons**:
- Not a container registry, it is a distribution accelerator
- Minimum deployment: 9+ pods (3x Manager, 3x Scheduler, 3x Seed Peer)
- Requires 60+ GB RAM minimum across the deployment
- Designed for 500+ node clusters, completely impractical for a 4-node homelab
- No image scanning

---

### Option 4: Wait for Harbor ARM64

Wait for the upstream goharbor project to ship multi-arch images.

**Pros**:
- No migration effort
- Harbor is CNCF Graduated with a large community

**Cons**:
- ARM64 PRs are still pending upstream with no clear timeline
- Cluster has no registry in the meantime (already broken for days)
- Harbor requires 7+ pods and 2-4 GB RAM even when working
- The hardcoded database password (D23 finding) would still need remediation

---

## Comparison Matrix

| Criteria | Zot | Distribution | Dragonfly | Wait for Harbor |
|---|---|---|---|---|
| CNCF Status | Sandbox | Sandbox | Graduated | Graduated |
| ARM64 / Pi CM4 | Native + Pi optimized | Native | Yes but impractical | No (amd64-only) |
| Image Scanning | Built-in Trivy | None | None | Built-in Trivy |
| Pull-Through Cache | Yes (Sync extension) | Yes (proxy mode) | Yes (P2P) | Yes |
| Web UI / Dashboard | Yes (built-in) | None | Yes (cluster mgmt) | Yes (full) |
| Resource Footprint | 1 pod, ~512Mi-1Gi | 1 pod, minimal | 9+ pods, 60+ GB | 7+ pods, 2-4 GB |
| Central Scanning Hub | Yes | No | No | Yes |
| Kyverno Compatible | Yes | Yes | No (not a registry) | Yes |
| Migration Effort | Medium | Low | High | None |

---

## Decision

**Option 1: Zot Registry** is the recommended replacement for the following reasons:

1. **Only option meeting all requirements**: Zot is the only CNCF registry project that provides ARM64 support, built-in scanning, pull-through caching, and a web dashboard simultaneously
2. **Resource-appropriate**: A single pod with no external database or Redis fits the CM4 constraints far better than Harbor's 7+ pod deployment
3. **Validated technology**: Harbor Satellite (Harbor's own edge/IoT solution) chose to embed Zot as its registry engine, which is strong validation from the Harbor project itself
4. **OCI-native**: Built for the OCI specification from the ground up, with native support for OCI artifacts, signatures, and SBOMs
5. **Raspberry Pi optimized**: The `commit` storage driver is specifically designed for memory-constrained environments like the CM4

This decision supersedes D20 (Harbor ARM64 migration via Bitnami).

---

## Implementation Plan

### Phase 1: Deploy Zot

1. Add Zot Helm chart entry to `appsets/applications-chart-application-set.yaml`
2. Create `applications/zot/values.yaml` with:
   - Longhorn storage (10Gi, matching current Harbor allocation)
   - Resource limits appropriate for CM4 (~300m CPU, 512Mi-1Gi RAM)
   - Web UI enabled
   - ARM64-compatible configuration
3. Create `applications/zot/manifests/` with namespace, ingress (`zot.mrcontainer.nl`), cert, and Tailscale Ingress
4. Verify Zot is running and the web UI is accessible

### Phase 2: Configure Pull-Through Caching

Configure the Sync extension with `onDemand: true` for all 7 upstream registries currently proxied through Harbor:

| Sync Prefix | Upstream Registry |
|---|---|
| `dockerhub` | `https://registry-1.docker.io` |
| `ghcr` | `https://ghcr.io` |
| `quay` | `https://quay.io` |
| `lscr` | `https://lscr.io` |
| `k8s` | `https://registry.k8s.io` |
| `gcr` | `https://gcr.io` |
| `nvcr` | `https://nvcr.io` |

### Phase 3: Enable Image Scanning

1. Enable the built-in Trivy search extension in Zot configuration
2. Configure automatic background scanning interval
3. Verify CVE results appear in the web UI

### Phase 4: Migrate Kyverno Policies

1. Update Kyverno ClusterPolicy (D10) to reference Zot registry URL instead of Harbor
2. Update image pull patterns from `harbor.mrcontainer.nl/<project>/` to `zot.mrcontainer.nl/<prefix>/`
3. Test that Kyverno correctly enforces the new image source policy

### Phase 5: Remove Harbor

1. Remove Harbor entry from `appsets/applications-chart-application-set.yaml`
2. Remove `applications/harbor/` directory (values.yaml, manifests)
3. Clean up Harbor PVCs from the cluster
4. Update Homepage dashboard to replace Harbor with Zot

---

## Dependencies

| ADR | Relationship |
|---|---|
| D9 | Supersedes: Harbor as container registry/proxy cache |
| D10 | Requires update: Kyverno policies must point to Zot |
| D17 | No dependency: Zot has no external database requirement |
| D20 | Supersedes: Harbor ARM64 migration is no longer viable |
| D23 | Resolves: Eliminates Harbor's hardcoded database password |

---

## References

- [Zot Registry](https://zotregistry.dev/)
- [Zot GitHub](https://github.com/project-zot/zot)
- [Zot CNCF Project Page](https://www.cncf.io/projects/zot/)
- [Zot Sync Extension (Pull-Through Cache)](https://zotregistry.dev/v2.1.10/articles/mirroring/)
- [Zot CVE Scanning](https://zotregistry.dev/v2.1.11/articles/cve-scanning/)
- [Harbor Satellite (embeds Zot)](https://github.com/container-registry/harbor-satellite)
- D9: Use Harbor as container registry / proxy cache
- D10: Use Kyverno to enforce registry usage
- D20: Harbor Migration to ARM64 (superseded)
