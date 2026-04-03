# Homelab Turing Pi HA Kubernetes

This guide walks through installing a High-Availability Talos Linux Kubernetes cluster on a Turing Pi v2.5 using four Raspberry Pi CM4 modules. Three nodes run as control plane, one as a worker.

## Prerequisites

* **Hardware**: Turing Pi v2.5 with 4 CM4 slots, ITX case, power supply
* **BMC access**: `tpi` CLI available and SSH access to the BMC
* **Network**: Static IPs assigned per node (configured in machine config)
* **Tools**: `talosctl` and `kubectl` installed on your local machine
* **Talos image**: Raspberry Pi ARM64 image from [factory.talos.dev](https://factory.talos.dev)

## Node Layout

| Node | IP              | Role          |
|------|-----------------|---------------|
| node1 | 192.168.68.101 | Control plane |
| node2 | 192.168.68.102 | Control plane |
| node3 | 192.168.68.103 | Control plane |
| node4 | 192.168.68.104 | Worker        |

## 1. Flash Talos Image

Flash the Talos Raspberry Pi image to each node via the Turing Pi BMC. Run these from your local machine over SSH to the BMC, or directly on the BMC.

Download the Talos image for Raspberry Pi:

```bash
# Download the Talos metal image for ARM64 (Raspberry Pi)
curl -LO https://factory.talos.dev/image/factory/metal-arm64.raw.xz
```

Flash each node slot using the `tpi` CLI:

```bash
tpi flash -n 1 -i metal-arm64.raw.xz
tpi flash -n 2 -i metal-arm64.raw.xz
tpi flash -n 3 -i metal-arm64.raw.xz
tpi flash -n 4 -i metal-arm64.raw.xz
```

Power cycle each node after flashing:

```bash
tpi power -n 1 off && tpi power -n 1 on
tpi power -n 2 off && tpi power -n 2 on
tpi power -n 3 off && tpi power -n 3 on
tpi power -n 4 off && tpi power -n 4 on
```

Nodes will boot into Talos maintenance mode, waiting for configuration to be applied.

## 2. Generate Cluster Configuration

Generate the base machine configs using `talosctl`. The VIP (`192.168.68.100`) provides a stable endpoint for the HA control plane:

```bash
talosctl gen config homelab-turing https://192.168.68.100:6443 \
  --output-dir _out
```

This produces:
- `_out/controlplane.yaml` — base config for control plane nodes
- `_out/worker.yaml` — base config for worker nodes
- `_out/talosconfig` — client config for `talosctl`

## 3. Customize Node Configs

Each node needs its own machine config with a static IP, hostname, and (for control plane nodes) a Virtual IP. Create a patch file per node and merge it into the base config.

### Control plane patch (repeat for node1/node2/node3, adjusting hostname and address)

**patch-node1.yaml**:
```yaml
machine:
  network:
    hostname: node1
    interfaces:
      - interface: eth0
        dhcp: false
        addresses:
          - 192.168.68.101/24
        routes:
          - network: 0.0.0.0/0
            gateway: 192.168.68.1
        vip:
          ip: 192.168.68.100
    nameservers:
      - 192.168.68.1
      - 8.8.8.8
```

Apply the patch to generate a per-node config:

```bash
talosctl machineconfig patch _out/controlplane.yaml \
  --patch @patch-node1.yaml \
  --output _out/node1.yaml

talosctl machineconfig patch _out/controlplane.yaml \
  --patch @patch-node2.yaml \
  --output _out/node2.yaml

talosctl machineconfig patch _out/controlplane.yaml \
  --patch @patch-node3.yaml \
  --output _out/node3.yaml
```

### Worker patch

**patch-node4.yaml**:
```yaml
machine:
  network:
    hostname: node4
    interfaces:
      - interface: eth0
        dhcp: false
        addresses:
          - 192.168.68.104/24
        routes:
          - network: 0.0.0.0/0
            gateway: 192.168.68.1
    nameservers:
      - 192.168.68.1
      - 8.8.8.8
  disks:
    - device: /dev/sda
      partitions:
        - mountpoint: /var/mnt/storage
```

> The `disks` section mounts the USB SSD on node4 for Jellyfin media storage.

```bash
talosctl machineconfig patch _out/worker.yaml \
  --patch @patch-node4.yaml \
  --output _out/node4.yaml
```

## 4. Apply Configuration

Apply each node's config while nodes are in maintenance mode (`--insecure` is needed before TLS is established):

```bash
talosctl apply-config --insecure --nodes 192.168.68.101 --file _out/node1.yaml
talosctl apply-config --insecure --nodes 192.168.68.102 --file _out/node2.yaml
talosctl apply-config --insecure --nodes 192.168.68.103 --file _out/node3.yaml
talosctl apply-config --insecure --nodes 192.168.68.104 --file _out/node4.yaml
```

Nodes will reboot and apply their configuration. Wait for them to come back up before continuing.

## 5. Bootstrap the Cluster

Bootstrap etcd on the first control plane node. This only needs to be done once:

```bash
talosctl bootstrap \
  --nodes 192.168.68.101 \
  --endpoints 192.168.68.101 \
  --talosconfig _out/talosconfig
```

Wait for the control plane to become healthy before proceeding.

## 6. Get kubeconfig

```bash
talosctl kubeconfig \
  --nodes 192.168.68.101 \
  --endpoints 192.168.68.101 \
  --talosconfig _out/talosconfig
```

Verify the cluster is up:

```bash
kubectl get nodes
```

All four nodes should appear. node1–node3 as control-plane, node4 as worker.

## 7. Install ArgoCD

```bash
helm repo add argo https://argoproj.github.io/argo-helm
helm repo update
helm install argocd argo/argo-cd --namespace argocd --create-namespace
```

After installation, connect ArgoCD to this Git repository through the ArgoCD UI or CLI. Then apply the ApplicationSets to start syncing everything:

```bash
kubectl apply -f core-components-chart-application-set.yaml
kubectl apply -f core-components-manifest-application-set.yaml
kubectl apply -f applications-chart-application-set.yaml
kubectl apply -f applications-manifest-application-set.yaml
```

ArgoCD will discover and deploy all components from this repository.

## Useful talosctl Commands

```bash
# Check node health
talosctl health --nodes 192.168.68.101 --endpoints 192.168.68.101

# View logs
talosctl logs --nodes 192.168.68.101 kubelet

# Reboot a node
talosctl reboot --nodes 192.168.68.101

# Upgrade Talos on a node
talosctl upgrade --nodes 192.168.68.101 --image ghcr.io/siderolabs/installer:<version>

# Upgrade Kubernetes
talosctl upgrade-k8s --nodes 192.168.68.101 --to <version>
```
