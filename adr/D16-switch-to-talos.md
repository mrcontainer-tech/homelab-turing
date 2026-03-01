# D16: Switch Kubernetes Distribution from K3s to Talos

**Date**: 2026-03-01
**Status**: Accepted
**Decision**: Replace K3s with Talos Linux as the cluster OS and Kubernetes distribution
**Context**: K3s was working, but its general-purpose OS model no longer aligned with how the cluster was actually being operated

---

## Problem Statement

The cluster has been running K3s on top of Raspberry Pi OS. While K3s works well, a few things became apparent over time:

- SSH access to nodes was never actually used. All operations went through GitOps (ArgoCD) or the Kubernetes API
- Having a full Linux OS (shell, SSH, package manager) on each node added unnecessary attack surface
- Node configuration (networking, mounts, kernel settings) was managed ad hoc, outside of GitOps
- K3s couples the OS and Kubernetes runtime in a way that makes it harder to manage nodes declaratively

The gap between "how the cluster is operated" (fully declarative, GitOps-driven) and "what the nodes look like" (general-purpose Linux with manual config) was becoming a point of friction.

## Goals

- Align node configuration with the GitOps model — everything declarative, nothing manual
- Reduce the attack surface on each node
- Keep full compatibility with the existing ArgoCD-based application stack
- Support current hardware (Raspberry Pi CM4) and potential future hardware (Turing RK1)

---

## Options Considered

### Option 1: Stay with K3s

Keep the existing setup. Address pain points incrementally (e.g., disable SSH, harden the OS manually).

**Pros**:
- No migration effort
- Familiar, well-documented
- Large community and ecosystem

**Cons**:
- Node-level config still managed imperatively
- Full Linux OS on each node, even though it's not used
- SSH surface remains even if access is disabled
- Doesn't improve the declarative story at the OS level

---

### Option 2: Talos Linux

Replace the OS entirely with Talos — a minimal, immutable Linux purpose-built for Kubernetes. No shell, no SSH, no package manager. The entire node is configured through a declarative YAML API (`talosctl`).

**Pros**:
- Immutable OS: no drift, no manual changes possible
- Dramatically reduced attack surface (no SSH, no shell, no extra services — only the kernel, containerd, and Kubernetes components)
- Node configuration is YAML, fully version-controllable and GitOps-compatible
- `talosctl` API replaces SSH for all node operations
- Supports Raspberry Pi CM4 and Turing RK1 (forward-compatible with potential hardware upgrades)
- Kubernetes upgrades are also managed declaratively via machine config

**Cons**:
- No SSH: debugging requires `talosctl` and thinking differently about node access
- Initial migration requires reinstalling all nodes
- Smaller community than K3s, though growing quickly
- Some tooling (e.g., host path mounts for USB storage) requires explicit machine config entries

---

### Option 3: MicroK8s

Lightweight Kubernetes distribution from Canonical, similar weight class to K3s.

**Pros**:
- Simple install
- Good addon ecosystem

**Cons**:
- Same general-purpose OS problem as K3s
- Doesn't improve declarative node management
- Less alignment with the CNCF ecosystem direction

---

## Comparison Matrix

| Criteria | K3s | Talos | MicroK8s |
|----------|-----|-------|----------|
| Declarative node config | No | Yes | No |
| Immutable OS | No | Yes | No |
| SSH required | Yes | No | Yes |
| Attack surface | Medium | Minimal | Medium |
| ARM64 / RPi support | Yes | Yes | Yes |
| Turing RK1 support | Unknown | Yes | Unknown |
| GitOps compatible | Yes | Yes | Yes |
| Migration effort | None | High (reinstall) | High |

---

## Decision

**Option 2: Talos Linux** was chosen for the following reasons:

1. **Already operated without SSH**: The cluster was never managed through SSH in practice. Talos formalises this operational model rather than leaving unused surface exposed
2. **Declarative all the way down**: Talos machine configs are YAML, stored in Git, and applied via `talosctl`. This extends the GitOps model from the application layer down to the OS layer
3. **Immutability by design**: Nodes cannot drift from their desired state. There is nothing to configure manually, nothing to install, nothing to break
4. **Security posture**: No shell, no SSH, no package manager. The attack surface on each node is as small as it can possibly be while still running Kubernetes
5. **Hardware flexibility**: Talos supports both the current Raspberry Pi CM4 modules and the Turing RK1 compute modules, leaving room to expand the cluster without changing the operating model

The migration cost (reinstalling all nodes) was accepted as a one-time investment for a significantly better long-term operational model.

---

## Implementation Notes

Node configuration is managed through Talos machine configs. USB-attached storage (used by Jellyfin on node4) requires an explicit disk mount in the machine config:

```yaml
machine:
  disks:
    - device: /dev/sda
      partitions:
        - mountpoint: /var/mnt/storage
```

This replaces the previous approach of relying on the host OS to mount the disk.

Cluster bootstrap is handled via `talosctl`:

```bash
talosctl gen config my-homelab https://<control-plane-ip>:6443
talosctl apply-config --insecure --nodes <node-ip> --file controlplane.yaml
talosctl bootstrap --nodes <control-plane-ip> --endpoints <control-plane-ip>
talosctl kubeconfig --nodes <control-plane-ip> --endpoints <control-plane-ip>
```

The ArgoCD-based application stack required no changes — Talos is a conformant Kubernetes distribution.

---

## References

- [Talos Linux documentation](https://www.talos.dev/docs/)
- [Talos on Raspberry Pi](https://www.talos.dev/latest/talos-guides/install/single-board-computers/rpi_generic/)
- [Talos on Turing Pi](https://www.talos.dev/latest/talos-guides/install/single-board-computers/turing_pi2/)
- Blog post: "Talos for my Homelab" (2026-03-01)
