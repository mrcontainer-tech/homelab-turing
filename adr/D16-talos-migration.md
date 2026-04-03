# D16: Migration from k3s to Talos Linux

**Date**: 2026-04-03
**Status**: Proposed
**Decision**: Migrate from k3s + Raspberry Pi OS to Talos Linux
**Context**: Security hardening, reduced maintenance, and planned hardware expansion
**Supersedes**: D6 (Raspberry Pi OS over Talos), D7 (k3s as Kubernetes distro)

---

## Problem Statement

The cluster currently runs k3s on Raspberry Pi OS across four CM4 nodes. While this has served well as an initial setup, several concerns have emerged:

1. **Security surface**: Raspberry Pi OS is a full Linux distribution with SSH access, package manager, and mutable filesystem — all unnecessary attack surface for a Kubernetes-only workload
2. **Maintenance burden**: OS-level updates (apt), security patches, and configuration drift require manual intervention per node
3. **Hardware expansion**: Planned additions of NVIDIA Jetson Orin NX and Turing RK1 modules make this a natural inflection point to rethink the OS layer

The original decision (D6) to use Raspberry Pi OS was driven by wanting SSH access for troubleshooting. With more Kubernetes experience and confidence in API-driven management, this trade-off is worth revisiting.

## Goals

- Immutable, minimal OS with reduced attack surface
- API-driven node management (no SSH)
- Declarative configuration that fits GitOps workflows
- Support for current hardware (CM4) and planned hardware (Jetson Orin NX, Turing RK1)
- Maintain Longhorn storage and existing workloads

---

## Options Considered

### Option 1: Stay on k3s + Raspberry Pi OS

Keep the current setup unchanged.

**Pros**:
- Known, stable, working setup
- SSH access for troubleshooting
- k3s is lightweight and resource-efficient
- Large community and documentation

**Cons**:
- Mutable OS allows configuration drift
- Manual OS maintenance per node
- Broader attack surface
- No immutability guarantees

---

### Option 2: Talos Linux

Replace Raspberry Pi OS with Talos Linux — a minimal, immutable, API-managed OS purpose-built for Kubernetes.

**Architecture**:
```
talosctl → Talos API (gRPC) → Immutable OS → Vanilla Kubernetes
```

**Pros**:
- Immutable filesystem, no SSH, minimal attack surface (12 binaries in PATH)
- API-driven management via talosctl — fits GitOps naturally
- Machine configs are single YAML files, version-controllable in Git
- Runs vanilla upstream Kubernetes (not k3s)
- Supports all target hardware:
  - **Raspberry Pi CM4**: Community-supported via `metal-rpi_generic` image, works on Turing Pi v2.5
  - **NVIDIA Jetson Orin NX**: Officially supported with Jetson-specific images and GPU access
  - **Turing RK1 (RK3588)**: Officially documented by Sidero Labs
- GitOps tooling: talhelper (declarative config), talm (GitOps patches), SOPS for secrets
- Automatic, safe upgrades with rollback

**Cons**:
- No SSH access — debugging is API-only (talosctl dmesg, logs, etc.)
- Learning curve for talosctl and machine config management
- Vanilla Kubernetes is heavier than k3s — slightly more resource usage
- Longhorn requires system extensions (iscsi-tools, util-linux-tools) and careful upgrade handling
- CM4 support is community-maintained, not officially certified
- RK1 NPU/GPU acceleration not available through Talos (kernel/driver limitation)

---

### Option 3: Flatcar Container Linux

Replace Raspberry Pi OS with Flatcar — an immutable, container-optimized Linux distribution.

**Pros**:
- Immutable /usr, auto-updates
- SSH still available for emergencies
- Supports multiple container orchestrators
- Larger binary set (2300+) for flexibility

**Cons**:
- Less strict immutability than Talos (runtime modifications still possible)
- Not Kubernetes-specific — more general-purpose
- Less ARM64 SBC support than Talos
- No specific Turing RK1 or Jetson documentation

---

### Option 4: Bottlerocket (AWS)

AWS-maintained container-optimized OS.

**Pros**:
- Immutable, minimal, purpose-built for containers

**Cons**:
- Designed for AWS/EKS — poor fit for bare-metal homelab
- No ARM64 SBC support
- Limited community for non-AWS use cases

---

## Comparison Matrix

| Criteria                  | k3s + Pi OS | Talos        | Flatcar      | Bottlerocket |
|---------------------------|-------------|--------------|--------------|--------------|
| Immutability              | None        | Full         | Partial      | Full         |
| Attack Surface            | Large       | Minimal      | Medium       | Minimal      |
| SSH Access                | Yes         | No (API)     | Yes          | Limited      |
| CM4 Support               | Excellent   | Community    | Limited      | No           |
| Jetson Orin NX Support    | Manual      | Official     | Unknown      | No           |
| Turing RK1 Support        | Manual      | Official     | Unknown      | No           |
| GitOps Fit                | Partial     | Excellent    | Good         | Good         |
| Maintenance Burden        | High        | Low          | Medium       | Low          |
| Kubernetes Distribution   | k3s         | Vanilla K8s  | Any          | K8s/EKS      |
| Longhorn Compatibility    | Native      | With extensions | Native    | Limited      |
| Resource Usage            | Low (k3s)   | Medium       | Medium       | Medium       |

---

## Decision

**Option 2: Talos Linux** is recommended for the following reasons:

1. **Immutability and security**: Minimal OS with no SSH eliminates entire categories of security concerns — aligns with the project's evolution from learning platform to production-grade homelab
2. **Maintenance reduction**: API-driven management and automatic upgrades replace manual per-node OS maintenance
3. **Hardware support**: Best coverage across all planned hardware — CM4 (community), Jetson Orin NX (official), Turing RK1 (official)
4. **GitOps alignment**: Machine configs as YAML in Git, managed via talhelper/SOPS, fits the existing ArgoCD-driven workflow perfectly
5. **Natural timing**: Node5 decommissioning and planned hardware additions create a clean migration window

The loss of SSH access is an acceptable trade-off given talosctl's diagnostic capabilities (dmesg, logs, services, containers) and the project's maturity beyond the initial setup phase.

---

## Migration Considerations

### Longhorn Storage
- Requires system extensions: `siderolabs/iscsi-tools` and `siderolabs/util-linux-tools`
- Kernel modules needed: nbd, iscsi_tcp, iscsi_generic, configfs
- **Critical**: Use `--preserve` flag during node upgrades to protect /var/lib/longhorn data
- Pod security baseline profile adjustments may be needed

### k3s to Vanilla Kubernetes
- Talos runs vanilla upstream Kubernetes, not k3s
- Traefik (k3s default) will need explicit installation
- ServiceLB (k3s built-in) replaced by MetalLB (already in use)
- No impact on ArgoCD, Longhorn, or application workloads

### Migration Strategy
- Migrate one node at a time (rolling replacement)
- Start with a worker node (node4) to validate workload compatibility
- Control plane nodes (node1-3) migrated after worker validation
- Back up all Longhorn volumes to S3 before migration

---

## Implementation Plan

1. **Preparation**:
   - Generate Talos machine configs with talhelper
   - Store configs in Git (encrypted secrets via SOPS)
   - Prepare Talos images for CM4 via Image Factory (include Longhorn extensions)
   - Full Longhorn S3 backup

2. **Worker migration** (node4):
   - Drain node4 workloads
   - Flash Talos image to CM4
   - Apply machine config via talosctl
   - Validate Longhorn, workloads, and ArgoCD sync

3. **Control plane migration** (node1-3):
   - Migrate one control plane node at a time
   - Validate etcd health after each migration
   - Ensure ArgoCD and cluster services remain operational

4. **Post-migration**:
   - Install Traefik via Helm (no longer bundled with k3s)
   - Validate all ApplicationSets and workloads
   - Update cluster documentation (TURING-INSTALL.md, README.md)
   - Update CLAUDE.md with Talos-specific commands

5. **New hardware**:
   - Add Jetson Orin NX with Talos Jetson image
   - Add Turing RK1 with Talos RK1 image
   - Update machine configs in Git

---

## References

- [Talos Linux Documentation](https://www.talos.dev/)
- [Talos on Raspberry Pi](https://www.talos.dev/v1.10/talos-guides/install/single-board-computers/rpi_generic/)
- [Talos on Turing RK1](https://docs.siderolabs.com/talos/v1.9/platform-specific-installations/single-board-computers/turing_rk1)
- [Talos on NVIDIA Jetson](https://www.talos.dev/v1.10/talos-guides/install/single-board-computers/jetson_sbc/)
- [Longhorn on Talos](https://longhorn.io/docs/1.7.2/advanced-resources/os-distro-specific/talos-linux-support/)
- [talhelper - Declarative Talos Config](https://github.com/budimanjojo/talhelper)
- [K3s vs Talos Linux](https://www.siderolabs.com/blog/talos-linux-vs-k3s/)
- [Talos vs Flatcar](https://www.siderolabs.com/blog/talos-linux-vs-flatcar/)
- D6: Use Raspberry Pi OS over Talos (superseded)
- D7: Use k3s as Kubernetes distro (superseded)
